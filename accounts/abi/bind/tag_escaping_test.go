// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package bind

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/accounts/abi"
)

func TestBindEscapesGeneratedABIFieldTags(t *testing.T) {
	const contractABI = `[
		{"inputs":[{"components":[
			{"name":"tuple_value","type":"uint512"}
		],"internalType":"struct Escaped.Record","name":"record","type":"tuple"}],"name":"store","outputs":[],"stateMutability":"nonpayable","type":"function"},
		{"anonymous":false,"inputs":[
			{"indexed":true,"name":"event\"quote","type":"uint512"},
			{"indexed":false,"name":"event\\slash","type":"bytes64"},
			{"indexed":false,"name":"event\u0001control","type":"address"}
		],"name":"Observed","type":"event"}
	]`
	code, err := Bind([]string{"EscapedTags"}, []string{contractABI}, []string{""}, nil, "bindtest", nil, nil)
	if err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "escaped_tags.go", code, parser.AllErrors|parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("generated escaped-tag binding does not parse: %v\n%s", err, code)
	}
	if diagnostics := generatedDeclarationErrors(fset, file); len(diagnostics) != 0 {
		t.Fatalf("generated escaped-tag binding has declaration conflicts: %v", diagnostics)
	}

	tags := make(map[string]int)
	var generatedABI string
	ast.Inspect(file, func(node ast.Node) bool {
		if valueSpec, ok := node.(*ast.ValueSpec); ok && len(valueSpec.Names) == 1 && valueSpec.Names[0].Name == "EscapedTagsMetaData" && len(valueSpec.Values) == 1 {
			address, ok := valueSpec.Values[0].(*ast.UnaryExpr)
			if !ok {
				t.Fatalf("generated metadata value has type %T, want address expression", valueSpec.Values[0])
			}
			literal, ok := address.X.(*ast.CompositeLit)
			if !ok {
				t.Fatalf("generated metadata address contains %T, want composite literal", address.X)
			}
			for _, element := range literal.Elts {
				pair, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, keyOK := pair.Key.(*ast.Ident)
				value, valueOK := pair.Value.(*ast.BasicLit)
				if !keyOK || key.Name != "ABI" || !valueOK {
					continue
				}
				generatedABI, err = strconv.Unquote(value.Value)
				if err != nil {
					t.Fatalf("unquote generated metadata ABI: %v", err)
				}
			}
		}
		structure, ok := node.(*ast.StructType)
		if !ok || structure.Fields == nil {
			return true
		}
		for _, field := range structure.Fields.List {
			if field.Tag == nil {
				continue
			}
			runtimeTag, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				t.Fatalf("unquote generated struct tag %s: %v", field.Tag.Value, err)
			}
			name, ok := reflect.StructTag(runtimeTag).Lookup("abi")
			if !ok {
				t.Fatalf("generated struct tag %q has no valid abi value", runtimeTag)
			}
			tags[name]++
		}
		return true
	})

	for _, name := range []string{
		"tuple_value",
		"event\"quote", "event\\slash", "event\x01control",
	} {
		if tags[name] != 1 {
			t.Errorf("generated ABI tag %q occurred %d times, want once; tags=%#v", name, tags[name], tags)
		}
	}
	if generatedABI == "" {
		t.Fatal("generated metadata ABI was not found")
	}
	parsed, err := abi.JSON(strings.NewReader(generatedABI))
	if err != nil {
		t.Fatalf("generated metadata ABI does not parse: %v\n%s", err, generatedABI)
	}
	eventInputs := parsed.Events["Observed"].Inputs
	for i, want := range []string{"event\"quote", "event\\slash", "event\x01control"} {
		if eventInputs[i].Name != want {
			t.Errorf("metadata event input %d name = %q, want %q", i, eventInputs[i].Name, want)
		}
	}
}
