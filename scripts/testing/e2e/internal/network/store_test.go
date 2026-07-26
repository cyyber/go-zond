// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

func TestNetworkDirectory(t *testing.T) {
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
	info, err := os.Stat(networkDir)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		t.Fatalf("%s mode = %v, err=%v", networkDir, info.Mode(), err)
	}
}

func TestOwnershipRequiresExactUUIDForDestruction(t *testing.T) {
	networkDir := t.TempDir()
	enclave := kurtosis.EnclaveRef{Name: enclaveNamePrefix(networkDir) + "attempt"}
	if err := validateOwnershipDirectory(networkDir, enclave); err == nil ||
		!strings.Contains(err.Error(), "UUID") {
		t.Fatalf("incomplete ownership error = %v", err)
	}
	enclave.UUID = strings.Repeat("a", 32)
	if err := validateOwnershipDirectory(networkDir, enclave); err != nil {
		t.Fatalf("valid ownership = %+v, err=%v", enclave, err)
	}
}

func TestOwnershipPersistsOneExclusiveExactIdentity(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	record := kurtosis.EnclaveRef{
		Name: enclaveNamePrefix(networkDir) + "0123456789abcdef0123456789abcdef",
		UUID: strings.Repeat("a", 32),
	}
	if err := createOwnership(networkDir, record); err != nil {
		t.Fatal(err)
	}
	if err := createOwnership(networkDir, record); !errors.Is(err, os.ErrExist) {
		t.Fatalf("duplicate ownership error = %v", err)
	}
	loaded, err := loadOwnership(networkDir)
	if err != nil || loaded != record {
		t.Fatalf("ownership=%+v err=%v", loaded, err)
	}
}

func TestOwnershipCannotBeCopiedToAnotherNetworkDirectory(t *testing.T) {
	source, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	destination, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	record := kurtosis.EnclaveRef{
		Name: enclaveNamePrefix(source) + "0123456789abcdef0123456789abcdef",
		UUID: strings.Repeat("a", 32),
	}
	if err := createOwnership(source, record); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(ownershipPath(source))
	if err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(ownershipPath(destination), payload); err != nil {
		t.Fatal(err)
	}
	if _, err := loadOwnership(destination); err == nil ||
		!strings.Contains(err.Error(), "does not belong") {
		t.Fatalf("copied ownership error = %v", err)
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
