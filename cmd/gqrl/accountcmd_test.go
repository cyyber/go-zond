// Copyright 2016 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/theQRL/go-qrl/accounts/keystore"
	"github.com/theQRL/go-qrl/common"
)

// These tests are 'smoke tests' for the account related
// subcommands and flags.
//
// Freshly-generated ML-DSA-87 keystore accounts are written into the
// temporary datadir at runtime (using live keygen) so the on-disk
// testdata tree doesn't need to carry address-format-specific blobs.

// tmpDatadirWithKeystore creates a datadir containing three freshly-generated
// keystore accounts with filenames "UTC--...--<addr>", "aaa", and "zzz". The
// returned addresses are in the order the keystore sorts them (by byte-wise
// filename comparison: 'U' < 'a' < 'z').
func tmpDatadirWithKeystore(t *testing.T) (datadir string, addrs [3]common.Address, utcFilename string) {
	t.Helper()
	datadir = t.TempDir()
	keydir := filepath.Join(datadir, "keystore")
	if err := os.MkdirAll(keydir, 0700); err != nil {
		t.Fatalf("mkdir keystore: %v", err)
	}
	ks := keystore.NewKeyStore(keydir, keystore.LightArgon2idT, keystore.LightArgon2idM, keystore.LightArgon2idP)
	mk := func(basename string) (accountURL string, addr common.Address) {
		acc, err := ks.NewAccount("foobar")
		if err != nil {
			t.Fatalf("NewAccount(%q): %v", basename, err)
		}
		target := filepath.Join(keydir, basename)
		if err := os.Rename(acc.URL.Path, target); err != nil {
			t.Fatalf("rename %q → %q: %v", acc.URL.Path, target, err)
		}
		return target, acc.Address
	}
	utcBase := "UTC--2025-11-06T07-34-54.273240000Z--keystore-test-0"
	_, addrs[0] = mk(utcBase)
	_, addrs[1] = mk("aaa")
	_, addrs[2] = mk("zzz")
	utcFilename = utcBase
	return
}

func TestAccountListEmpty(t *testing.T) {
	t.Parallel()
	gqrl := runGqrl(t, "account", "list")
	gqrl.ExpectExit()
}

func TestAccountList(t *testing.T) {
	t.Parallel()
	datadir, addrs, utcFilename := tmpDatadirWithKeystore(t)
	sep := "/"
	if runtime.GOOS == "windows" {
		sep = `\`
	}
	// Address list output uses the lower-case hex form; construct the
	// expected string dynamically from the addresses we just generated.
	lowerHex := func(a common.Address) string {
		return fmt.Sprintf("%#x", a)
	}
	want := fmt.Sprintf("\n"+
		"Account #0: {%s} keystore://{{.Datadir}}%skeystore%s%s\n"+
		"Account #1: {%s} keystore://{{.Datadir}}%skeystore%saaa\n"+
		"Account #2: {%s} keystore://{{.Datadir}}%skeystore%szzz\n",
		lowerHex(addrs[0]), sep, sep, utcFilename,
		lowerHex(addrs[1]), sep, sep,
		lowerHex(addrs[2]), sep, sep,
	)
	{
		gqrl := runGqrl(t, "account", "list", "--datadir", datadir)
		gqrl.Expect(want)
		gqrl.ExpectExit()
	}
	{
		gqrl := runGqrl(t, "--datadir", datadir, "account", "list")
		gqrl.Expect(want)
		gqrl.ExpectExit()
	}
}

func TestAccountNew(t *testing.T) {
	t.Parallel()
	gqrl := runGqrl(t, "account", "new", "--lightkdf")
	defer gqrl.ExpectExit()
	gqrl.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar"}}
Repeat password: {{.InputLine "foobar"}}

Your new key was generated
`)
	// QRL addresses are 64 bytes → 128 hex characters after the Q prefix.
	gqrl.ExpectRegexp(`
Public address of the key:   Q[0-9a-fA-F]{128}
Path of the secret key file: .*UTC--.+--Q[0-9a-f]{128}

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
`)
}

func TestAccountImport(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, seed, output string }{
		{
			name:   "correct account",
			seed:   "0100000123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdeffcad0b19bb29d4674531d6f115237e16",
			output: "Address: {Q958d36976b91586a10341cf20c7dfbcb122a106533cef625327b684878c1755196fd25156fc39b43291dce296aceea83d716e1ef1e0382a8984efc12185426e4}\n",
		},
		{
			name:   "invalid character",
			seed:   "0100000123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdeffcad0b19bb29d4674531d6f115237e161",
			output: "Fatal: Failed to restore wallet from file: invalid character '1' at end of key file\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			importAccountWithExpect(t, test.seed, test.output)
		})
	}
}

func TestAccountHelp(t *testing.T) {
	t.Parallel()
	gqrl := runGqrl(t, "account", "-h")
	gqrl.WaitExit()
	if have, want := gqrl.ExitStatus(), 0; have != want {
		t.Errorf("exit error, have %d want %d", have, want)
	}

	gqrl = runGqrl(t, "account", "import", "-h")
	gqrl.WaitExit()
	if have, want := gqrl.ExitStatus(), 0; have != want {
		t.Errorf("exit error, have %d want %d", have, want)
	}
}

func importAccountWithExpect(t *testing.T, seed string, expected string) {
	dir := t.TempDir()
	seedfile := filepath.Join(dir, "seed.txt")
	if err := os.WriteFile(seedfile, []byte(seed), 0600); err != nil {
		t.Error(err)
	}
	passwordFile := filepath.Join(dir, "password.txt")
	if err := os.WriteFile(passwordFile, []byte("foobar"), 0600); err != nil {
		t.Error(err)
	}
	gqrl := runGqrl(t, "--lightkdf", "account", "import", "-password", passwordFile, seedfile)
	defer gqrl.ExpectExit()
	gqrl.Expect(expected)
}

func TestAccountNewBadRepeat(t *testing.T) {
	t.Parallel()
	gqrl := runGqrl(t, "account", "new", "--lightkdf")
	defer gqrl.ExpectExit()
	gqrl.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "something"}}
Repeat password: {{.InputLine "something else"}}
Fatal: Passwords do not match
`)
}

func TestAccountUpdate(t *testing.T) {
	t.Parallel()
	datadir, addrs, _ := tmpDatadirWithKeystore(t)
	// Target the account backed by the "zzz" file. The CLI accepts the
	// lower-case hex form and the OLD-password prompt echoes the canonical
	// QIP-55 form of the same address.
	zzzAddr := addrs[2]
	zzzLower := fmt.Sprintf("%#x", zzzAddr)
	gqrl := runGqrl(t, "account", "update",
		"--datadir", datadir, "--lightkdf",
		zzzLower)
	defer gqrl.ExpectExit()
	gqrl.Expect(fmt.Sprintf(`
Please give a NEW password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Password: {{.InputLine "foobar_new"}}
Repeat password: {{.InputLine "foobar_new"}}
Please provide the OLD password for account %s | Attempt 1/3
Password: {{.InputLine "foobar"}}
`, zzzAddr.Hex()))
}
