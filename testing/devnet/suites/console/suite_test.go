// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"io/fs"
	"testing"
)

func TestParseSuiteResult(t *testing.T) {
	valid := []byte(
		`CONSOLE_E2E_PASS api`,
	)
	if err := parseSuiteResult("api", valid); err != nil {
		t.Fatal(err)
	}

	if err := parseSuiteResult("api", []byte(`CONSOLE_E2E_FAIL api`)); err == nil {
		t.Fatal("failed suite was accepted")
	}
}

func TestSuiteFixtures(t *testing.T) {
	for _, name := range append([]string{"harness"}, suiteNames...) {
		if _, err := fs.Stat(consoleFixtures, "testdata/console/"+name+".js"); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
