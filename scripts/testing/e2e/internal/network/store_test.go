// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNetworkDirectoryAndReadyMarker(t *testing.T) {
	parent, linkRoot := t.TempDir(), t.TempDir()
	link := filepath.Join(linkRoot, "parent")
	if err := os.Symlink(parent, link); err != nil {
		t.Fatal(err)
	}
	networkDir, err := ensureNetworkDirectory(filepath.Join(link, "network"))
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(networkDir) || strings.Contains(networkDir, linkRoot) {
		t.Fatalf("network directory was not canonicalized: %s", networkDir)
	}
	for _, path := range []string{networkDir, privatePath(networkDir)} {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
			t.Fatalf("%s mode = %v, err=%v", path, info.Mode(), err)
		}
	}
	if err := writeState(networkDir, State{Ready: true}); err != nil {
		t.Fatal(err)
	}
	state, err := loadState(networkDir)
	if err != nil || !state.Ready {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}

func TestOwnershipMovesFromExclusiveIntentToExactIdentity(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	record := OwnershipRecord{NetworkDir: networkDir, Name: "e2e"}
	if err := createOwnership(record); err != nil {
		t.Fatal(err)
	}
	if err := createOwnership(record); !errors.Is(err, os.ErrExist) {
		t.Fatalf("duplicate intent error = %v", err)
	}
	record.UUID = strings.Repeat("a", 32)
	if err := captureOwnership(record); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadOwnership(networkDir)
	if err != nil || loaded != record {
		t.Fatalf("ownership=%+v err=%v", loaded, err)
	}
}

func TestExclusiveWriteDoesNotReplaceExistingSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	if err := writeExclusive(path, []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(path, []byte("second\n")); !errors.Is(err, os.ErrExist) {
		t.Fatalf("replacement error = %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != "first\n" {
		t.Fatalf("exclusive content = %q, err=%v", contents, err)
	}
}

func TestNetworkDirectoryRejectsPrivateSymlink(t *testing.T) {
	networkDir := t.TempDir()
	if err := os.Symlink(t.TempDir(), privatePath(networkDir)); err != nil {
		t.Skipf("create private-directory symlink: %v", err)
	}
	if _, err := ensureNetworkDirectory(networkDir); err == nil ||
		!strings.Contains(err.Error(), "non-symlink") {
		t.Fatalf("private symlink error = %v", err)
	}
}
