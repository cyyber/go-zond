// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package compiler wraps the ABI compilation outputs.
package compiler

import (
	"encoding/json"
)

// --combined-output format
type hypcOutput struct {
	Contracts map[string]struct {
		BinRuntime            string `json:"bin-runtime"`
		SrcMapRuntime         string `json:"srcmap-runtime"`
		Bin, SrcMap, Metadata string
		Abi                   interface{}
		Devdoc                interface{}
		Userdoc               interface{}
		Hashes                map[string]string
	}
	Version string
}

// ParseCombinedJSON takes the direct output of a hypc --combined-output run and
// parses it into a map of string contract name to Contract structs. The
// provided source, language and compiler version, and compiler options are all
// passed through into the Contract structs.
//
// The hypc output is expected to contain ABI, source mapping, user docs, and dev docs.
//
// Returns an error if the JSON is malformed or missing data, or if the JSON
// embedded within the JSON is malformed.
func ParseCombinedJSON(combinedJSON []byte, source string, languageVersion string, compilerVersion string, compilerOptions string) (map[string]*Contract, error) {
	var output hypcOutput
	if err := json.Unmarshal(combinedJSON, &output); err != nil {
		return nil, err
	}
	// Compilation succeeded, assemble and return the contracts.
	contracts := make(map[string]*Contract)
	for name, info := range output.Contracts {
		contracts[name] = &Contract{
			Code:        "0x" + info.Bin,
			RuntimeCode: "0x" + info.BinRuntime,
			Hashes:      info.Hashes,
			Info: ContractInfo{
				Source:          source,
				Language:        "Hyperion",
				LanguageVersion: languageVersion,
				CompilerVersion: compilerVersion,
				CompilerOptions: compilerOptions,
				SrcMap:          info.SrcMap,
				SrcMapRuntime:   info.SrcMapRuntime,
				AbiDefinition:   info.Abi,
				UserDoc:         info.Userdoc,
				DeveloperDoc:    info.Devdoc,
				Metadata:        info.Metadata,
			},
		}
	}
	return contracts, nil
}
