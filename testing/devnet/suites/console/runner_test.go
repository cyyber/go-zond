// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSuiteResult(t *testing.T) {
	valid := []byte(
		`VM64_E2E_RESULT {"schema":1,"suite":"web3_sanity","status":"passed","passed":7,"failed":0,"total":7}`,
	)
	if err := parseSuiteResult("web3_sanity", valid); err != nil {
		t.Fatal(err)
	}

	invalid := []byte(
		`VM64_E2E_RESULT {"schema":1,"suite":"web3_sanity","status":"failed","passed":6,"failed":1,"total":7}`,
	)
	if err := parseSuiteResult("web3_sanity", invalid); err == nil {
		t.Fatal("failed suite was accepted")
	}
}

func TestSuiteFixtures(t *testing.T) {
	root, err := testdataRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range suiteNames {
		if _, err := os.Stat(filepath.Join(root, "console", name+".js")); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
