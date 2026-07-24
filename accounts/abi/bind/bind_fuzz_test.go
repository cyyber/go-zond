// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-qrl library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-qrl library. If not, see <http://www.gnu.org/licenses/>.

package bind

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"testing"
)

type generatedImportStub struct{}

func (generatedImportStub) Import(path string) (*types.Package, error) {
	name := strings.ReplaceAll(filepath.Base(path), "-", "_")
	return types.NewPackage(path, name), nil
}

// generatedDeclarationErrors applies Go's semantic checker to declarations in
// generated source. The imported packages are intentionally skeletal: this
// lane is looking for redeclarations and scope collisions, while
// TestGolangBindings performs the full module-aware go vet compilation.
func generatedDeclarationErrors(fset *token.FileSet, file *ast.File) []string {
	var diagnostics []string
	config := types.Config{
		Importer:         generatedImportStub{},
		IgnoreFuncBodies: true,
		Error: func(err error) {
			message := err.Error()
			for _, marker := range []string{"redeclared", "already declared", "field and method with the same name"} {
				if strings.Contains(message, marker) {
					diagnostics = append(diagnostics, message)
					break
				}
			}
		},
	}
	_, _ = config.Check(file.Name.Name, fset, []*ast.File{file}, nil)
	return diagnostics
}

func TestGeneratedDeclarationErrorsDetectSemanticDuplicates(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "duplicates.go", `package duplicates
		type Record struct { Raw int; Raw string }
		func duplicate(opts int, opts string) {}`, parser.AllErrors|parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("duplicate fixture must remain syntactically valid: %v", err)
	}
	if diagnostics := generatedDeclarationErrors(fset, file); len(diagnostics) < 2 {
		t.Fatalf("semantic checker missed duplicate field/parameter declarations: %v", diagnostics)
	}
}

// FuzzBindDeterministicAndParses checks that arbitrary ABI input cannot panic
// the Go binding generator. Whenever Bind accepts an ABI, generation must also
// be deterministic and produce syntactically valid Go.
func FuzzBindDeterministicAndParses(f *testing.F) {
	for _, seed := range []string{
		`[]`,
		`[{"inputs":[{"name":"account","type":"address"},{"name":"amount","type":"uint512"},{"name":"delta","type":"int512"},{"name":"salt","type":"bytes64"},{"name":"memo","type":"string"},{"name":"values","type":"uint512[]"},{"name":"fixedValues","type":"bytes64[2]"}],"name":"roundTrip","outputs":[{"name":"amount","type":"uint512"}],"stateMutability":"view","type":"function"}]`,
		`[{"anonymous":true,"inputs":[{"indexed":true,"name":"account","type":"address"},{"indexed":true,"name":"label","type":"string"},{"indexed":false,"name":"payload","type":"bytes64"}],"name":"AnonymousStored","type":"event"}]`,
		`[{"anonymous":false,"inputs":[{"indexed":true,"name":"text","type":"string"},{"indexed":true,"name":"blob","type":"bytes"},{"indexed":true,"name":"values","type":"uint512[]"},{"indexed":false,"name":"count","type":"uint512"}],"name":"CompositeTopics","type":"event"}]`,
		`[{"inputs":[{"components":[{"name":"owner","type":"address"},{"name":"amount","type":"uint512"}],"name":"record","type":"tuple"}],"name":"store","outputs":[{"components":[{"name":"owner","type":"address"},{"name":"amount","type":"uint512"}],"name":"record","type":"tuple"}],"stateMutability":"nonpayable","type":"function"}]`,
		`[{"inputs":[{"name":"empty","type":"uint512[0]"},{"name":"nested","type":"bytes64[][2]"},{"name":"dynamic","type":"address[2][]"}],"name":"collections","outputs":[],"stateMutability":"pure","type":"function"}]`,
		`[{"inputs":[{"name":"amount","type":"uint512"}],"name":"overloaded","outputs":[],"stateMutability":"view","type":"function"},{"inputs":[{"name":"account","type":"address"}],"name":"overloaded","outputs":[],"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"type","type":"address"},{"indexed":false,"name":"_value","type":"bytes64"}],"name":"overloaded","type":"event"}]`,
		`[{"constant":true,"inputs":[],"name":"legacyView","outputs":[{"name":"amount","type":"uint512"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"legacyPayable","outputs":[],"payable":true,"type":"function"}]`,
		`[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"stateMutability":"payable","type":"fallback"},{"stateMutability":"payable","type":"receive"},{"inputs":[{"name":"amount","type":"uint512"}],"name":"TooLarge","type":"error"}]`,
		`[{"inputs":[{"name":"callback","type":"function"}],"name":"configure","outputs":[],"stateMutability":"nonpayable","type":"function"}]`,
		reservedIdentifierABI,
		`[{`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, abiJSON []byte) {
		// Large generated sources add little coverage and make fuzz minimization
		// needlessly expensive. Nested and composite ABIs fit comfortably here.
		abiJSON = abiJSON[:min(len(abiJSON), 32<<10)]
		args := func() (string, error) {
			return Bind(
				[]string{"FuzzContract"},
				[]string{string(abiJSON)},
				[]string{""},
				nil,
				"fuzzbind",
				nil,
				nil,
			)
		}
		first, firstErr := args()
		second, secondErr := args()
		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("Bind returned inconsistent success for identical ABI: first=%v second=%v", firstErr, secondErr)
		}
		if firstErr != nil {
			if firstErr.Error() != secondErr.Error() {
				t.Fatalf("Bind returned nondeterministic errors for identical ABI: first=%q second=%q", firstErr, secondErr)
			}
			return
		}
		if first != second {
			t.Fatal("Bind generated different Go source for identical ABI")
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "fuzz_binding.go", first, parser.AllErrors|parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("Bind generated invalid Go: %v\n%s", err, first)
		}
		if diagnostics := generatedDeclarationErrors(fset, file); len(diagnostics) != 0 {
			t.Fatalf("Bind generated semantically invalid declarations: %v\n%s", diagnostics, first)
		}
	})
}
