// Copyright 2016 The go-ethereum Authors
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

// Package bind generates QRL contract Go bindings.
//
// Detailed usage document and tutorial available on the go-ethereum Wiki page:
// https://github.com/ethereum/go-ethereum/wiki/Native-DApps:-Go-bindings-to-Ethereum-contracts
package bind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"go/token"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/log"
)

var (
	intRegex = regexp.MustCompile(`(u)?int([0-9]*)`)
)

func isKeyWord(arg string) bool {
	switch arg {
	case "break":
	case "case":
	case "chan":
	case "const":
	case "continue":
	case "default":
	case "defer":
	case "else":
	case "fallthrough":
	case "for":
	case "func":
	case "go":
	case "goto":
	case "if":
	case "import":
	case "interface":
	case "iota":
	case "map":
	case "make":
	case "new":
	case "package":
	case "range":
	case "return":
	case "select":
	case "struct":
	case "switch":
	case "type":
	case "var":
	default:
		return false
	}

	return true
}

var predeclaredGoIdentifiers = map[string]bool{
	"any": true, "bool": true, "byte": true, "comparable": true,
	"complex64": true, "complex128": true, "error": true,
	"float32": true, "float64": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"rune": true, "string": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
	"true": true, "false": true, "iota": true, "nil": true,
	"append": true, "cap": true, "clear": true, "close": true, "complex": true,
	"copy": true, "delete": true, "imag": true, "len": true, "make": true,
	"max": true, "min": true, "new": true, "panic": true, "print": true,
	"println": true, "real": true, "recover": true,
}

func contractPackageNames(contract string) []string {
	return []string{
		contract,
		contract + "MetaData", contract + "ABI", contract + "Bin", contract + "FuncSigs",
		contract + "Caller", contract + "Transactor", contract + "Filterer",
		contract + "Session", contract + "CallerSession", contract + "TransactorSession",
		contract + "Raw", contract + "CallerRaw", contract + "TransactorRaw",
		"Deploy" + contract,
		"New" + contract, "New" + contract + "Caller", "New" + contract + "Transactor", "New" + contract + "Filterer",
		"bind" + contract,
	}
}

func identifierSet(names ...string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func generatedIdentifierBase(name, fallback string, index int) string {
	if name == "" || name == "_" || isKeyWord(name) || predeclaredGoIdentifiers[name] || !token.IsIdentifier(name) {
		return fmt.Sprintf("%s%d", fallback, index)
	}
	return name
}

func resolveGeneratedIdentifier(name, fallback string, index int, conflicts func(string) bool) string {
	base := generatedIdentifierBase(name, fallback, index)
	for suffix := -1; ; suffix++ {
		candidate := base
		if suffix >= 0 {
			candidate = fmt.Sprintf("%s%d", base, suffix)
		}
		if !conflicts(candidate) {
			return candidate
		}
	}
}

func normalizeConstructorInputs(inputs abi.Arguments, libs map[string]string, contract string) abi.Arguments {
	normalized := append(abi.Arguments(nil), inputs...)
	// Constructor arguments are in scope throughout the generated deployment
	// body. Reserve packages, locals and generated contract declarations used
	// there; otherwise a perfectly valid ABI name such as "bind" or
	// "ContractMetaData" shadows the identifier and produces uncompilable Go.
	used := identifierSet(
		"auth", "backend", "parsed", "err", "address", "tx", "contract",
		"bind", "common", "errors",
		contract, contract+"MetaData", contract+"Bin",
		contract+"Caller", contract+"Transactor", contract+"Filterer",
	)
	for _, library := range libs {
		if name := decapitalise(library); name != "" {
			used[name+"Addr"] = true
		}
		if name := abi.ToCamelCase(library); name != "" {
			used["Deploy"+name] = true
		}
	}
	if len(libs) > 0 {
		used["strings"] = true
	}
	for i := range normalized {
		name := resolveGeneratedIdentifier(normalized[i].Name, "arg", i, func(candidate string) bool { return used[candidate] })
		normalized[i].Name = name
		used[name] = true
	}
	return normalized
}

func normalizeMethodInputs(inputs abi.Arguments, contract string, outputCount int, usesABI bool) abi.Arguments {
	normalized := append(abi.Arguments(nil), inputs...)
	used := identifierSet("opts", "out", "err", "outstruct", "_"+contract)
	if usesABI {
		// Constant methods with outputs call abi.ConvertType in their generated
		// body, where an input parameter named "abi" would shadow the package.
		used["abi"] = true
	}
	for i := 0; i < outputCount; i++ {
		used[fmt.Sprintf("out%d", i)] = true
	}
	for i := range normalized {
		name := resolveGeneratedIdentifier(normalized[i].Name, "arg", i, func(candidate string) bool { return used[candidate] })
		normalized[i].Name = name
		used[name] = true
	}
	return normalized
}

func normalizeMethodOutputs(outputs abi.Arguments) abi.Arguments {
	normalized := append(abi.Arguments(nil), outputs...)
	used := make(map[string]bool)
	for i := range normalized {
		if normalized[i].Name == "" {
			continue
		}
		base := abi.ToCamelCase(normalized[i].Name)
		name := resolveGeneratedIdentifier(base, "Ret", i, func(candidate string) bool { return used[candidate] })
		normalized[i].Name = name
		used[name] = true
	}
	return normalized
}

func normalizeEventInputs(inputs abi.Arguments, contract string) abi.Arguments {
	normalized := append(abi.Arguments(nil), inputs...)
	usedVariables := identifierSet("opts", "sink", "logs", "sub", "err", "event", "quit", "log", "_"+contract)
	usedFields := identifierSet("Raw")
	for i := range normalized {
		indexed := normalized[i].Indexed
		name := resolveGeneratedIdentifier(normalized[i].Name, "arg", i, func(candidate string) bool {
			if usedVariables[candidate] || usedFields[abi.ToCamelCase(candidate)] {
				return true
			}
			return indexed && (usedVariables[candidate+"Rule"] || usedVariables[candidate+"Item"])
		})
		normalized[i].Name = name
		usedVariables[name] = true
		usedFields[abi.ToCamelCase(name)] = true
		if indexed {
			usedVariables[name+"Rule"] = true
			usedVariables[name+"Item"] = true
		}
	}
	return normalized
}

func resolveEventIdentifier(name, contract string, events, packageNames map[string]bool) string {
	if name == "" || name == "_" || !token.IsIdentifier(name) {
		name = "Event"
	}
	return resolveGeneratedIdentifier(name, "Event", 0, func(candidate string) bool {
		return events[candidate] || packageNames[contract+candidate] || packageNames[contract+candidate+"Iterator"]
	})
}

// Bind generates a Go wrapper around a contract ABI. This wrapper isn't meant
// to be used as is in client code, but rather as an intermediate struct which
// enforces compile time type safety and naming convention as opposed to having to
// manually maintain hard coded strings that break on runtime.
func Bind(types []string, abis []string, bytecodes []string, fsigs []map[string]string, pkg string, libs map[string]string, aliases map[string]string) (string, error) {
	if len(types) != len(abis) {
		return "", fmt.Errorf("binding type/ABI count mismatch: %d types, %d ABIs", len(types), len(abis))
	}
	if len(types) != len(bytecodes) {
		return "", fmt.Errorf("binding type/bytecode count mismatch: %d types, %d bytecodes", len(types), len(bytecodes))
	}
	var (
		// contracts is the map of each individual contract requested binding
		contracts = make(map[string]*tmplContract)

		// structs is the map of all redeclared structs shared by passed contracts.
		structs = make(map[string]*tmplStruct)

		// isLib is the map used to flag each encountered library as such
		isLib = make(map[string]struct{})

		// packageNames tracks every top-level identifier emitted by the template.
		// Tuple and event types share this namespace with contract wrappers,
		// metadata variables and constructors, so resolving only against other
		// tuples is insufficient.
		packageNames  = make(map[string]bool)
		contractTypes = make([]string, len(types))
	)
	for i, rawType := range types {
		contractType := abi.ToCamelCase(rawType)
		if contractType == "" || contractType == "_" || !token.IsIdentifier(contractType) {
			return "", fmt.Errorf("invalid binding type name %q", rawType)
		}
		contractTypes[i] = contractType
		for _, name := range contractPackageNames(contractType) {
			if packageNames[name] {
				return "", fmt.Errorf("binding type %q generates duplicate top-level identifier %q", rawType, name)
			}
			packageNames[name] = true
		}
	}
	for i := range types {
		// Parse the actual ABI to generate the binding for
		qrvmABI, err := abi.JSON(strings.NewReader(abis[i]))
		if err != nil {
			return "", err
		}
		if err := rejectFunctionTypes(types[i], qrvmABI); err != nil {
			return "", err
		}
		// Compact structural whitespace without changing whitespace embedded in
		// JSON strings such as internalType names or source-level identifiers.
		var compactABI bytes.Buffer
		if err := json.Compact(&compactABI, []byte(abis[i])); err != nil {
			return "", fmt.Errorf("compact ABI for %q: %w", types[i], err)
		}
		strippedABI := compactABI.String()

		// Extract the call and transact methods; events, struct definitions; and sort them alphabetically
		var (
			calls     = make(map[string]*tmplMethod)
			transacts = make(map[string]*tmplMethod)
			events    = make(map[string]*tmplEvent)
			fallback  *tmplMethod
			receive   *tmplMethod

			// identifiers are used to detect duplicated identifiers of functions
			// and events. For all calls, transacts and events, abigen will generate
			// corresponding bindings. However we have to ensure there is no
			// identifier collisions in the bindings of these categories.
			callIdentifiers     = make(map[string]bool)
			transactIdentifiers = make(map[string]bool)
			eventIdentifiers    = make(map[string]bool)
			sessionIdentifiers  = identifierSet("Contract", "CallOpts", "TransactOpts")
		)
		// Every ordinary method is emitted on the combined Session type, even
		// though caller and transactor methods otherwise have separate receiver
		// types. Reserve the generated session fields and special entrypoints in
		// that shared namespace so Bind cannot emit duplicate declarations.
		if qrvmABI.HasFallback() {
			sessionIdentifiers["Fallback"] = true
		}
		if qrvmABI.HasReceive() {
			sessionIdentifiers["Receive"] = true
		}

		constructor := qrvmABI.Constructor
		constructor.Inputs = normalizeConstructorInputs(constructor.Inputs, libs, contractTypes[i])
		for _, input := range constructor.Inputs {
			if hasStruct(input.Type) {
				bindStructType(input.Type, structs, packageNames)
			}
		}

		methodNames := make([]string, 0, len(qrvmABI.Methods))
		for name := range qrvmABI.Methods {
			methodNames = append(methodNames, name)
		}
		sort.Strings(methodNames)
		for _, name := range methodNames {
			original := qrvmABI.Methods[name]
			// Normalize the method for capital cases and non-anonymous inputs/outputs
			normalized := original
			normalizedName := abi.ToCamelCase(alias(aliases, original.Name))
			// Ensure there is no duplicated identifier
			var identifiers = callIdentifiers
			if !original.IsConstant() {
				identifiers = transactIdentifiers
			}
			// Name shouldn't start with a digit. It will make the generated code invalid.
			if len(normalizedName) > 0 && unicode.IsDigit(rune(normalizedName[0])) {
				normalizedName = fmt.Sprintf("M%s", normalizedName)
				normalizedName = abi.ResolveNameConflict(normalizedName, func(name string) bool {
					_, ok := identifiers[name]
					return ok
				})
			}
			if normalizedName == "" || !token.IsIdentifier(normalizedName) {
				return "", fmt.Errorf("method %q generates invalid Go identifier %q, use --alias for renaming", original.Name, normalizedName)
			}
			if identifiers[normalizedName] {
				return "", fmt.Errorf("duplicated identifier \"%s\"(normalized \"%s\"), use --alias for renaming", original.Name, normalizedName)
			}
			if sessionIdentifiers[normalizedName] {
				return "", fmt.Errorf("method %q generates duplicate session identifier %q, use --alias for renaming", original.Name, normalizedName)
			}
			identifiers[normalizedName] = true
			sessionIdentifiers[normalizedName] = true

			normalized.Name = normalizedName
			normalized.Inputs = make([]abi.Argument, len(original.Inputs))
			copy(normalized.Inputs, original.Inputs)
			normalized.Inputs = normalizeMethodInputs(
				normalized.Inputs,
				contractTypes[i],
				len(original.Outputs),
				original.IsConstant() && len(original.Outputs) > 0,
			)
			for _, input := range normalized.Inputs {
				if hasStruct(input.Type) {
					bindStructType(input.Type, structs, packageNames)
				}
			}
			normalized.Outputs = make([]abi.Argument, len(original.Outputs))
			copy(normalized.Outputs, original.Outputs)
			normalized.Outputs = normalizeMethodOutputs(normalized.Outputs)
			for _, output := range normalized.Outputs {
				if hasStruct(output.Type) {
					bindStructType(output.Type, structs, packageNames)
				}
			}
			// Append the methods to the call or transact lists
			if original.IsConstant() {
				calls[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
			} else {
				transacts[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
			}
		}
		eventNames := make([]string, 0, len(qrvmABI.Events))
		for name := range qrvmABI.Events {
			eventNames = append(eventNames, name)
		}
		sort.Strings(eventNames)
		for _, name := range eventNames {
			original := qrvmABI.Events[name]
			// Normalize the event for capital cases and non-anonymous outputs
			normalized := original

			// Ensure there is no duplicated identifier
			normalizedName := abi.ToCamelCase(alias(aliases, original.Name))
			// Name shouldn't start with a digit. It will make the generated code invalid.
			if len(normalizedName) > 0 && unicode.IsDigit(rune(normalizedName[0])) {
				normalizedName = fmt.Sprintf("E%s", normalizedName)
				normalizedName = abi.ResolveNameConflict(normalizedName, func(name string) bool {
					_, ok := eventIdentifiers[name]
					return ok
				})
			}
			normalizedName = resolveEventIdentifier(normalizedName, contractTypes[i], eventIdentifiers, packageNames)
			eventIdentifiers[normalizedName] = true
			packageNames[contractTypes[i]+normalizedName] = true
			packageNames[contractTypes[i]+normalizedName+"Iterator"] = true
			normalized.Name = normalizedName

			normalized.Inputs = make([]abi.Argument, len(original.Inputs))
			copy(normalized.Inputs, original.Inputs)
			normalized.Inputs = normalizeEventInputs(normalized.Inputs, contractTypes[i])
			for _, input := range normalized.Inputs {
				if hasStruct(input.Type) {
					bindStructType(input.Type, structs, packageNames)
				}
			}
			// Preserve the original ABI names alongside their normalized Go names.
			// Generated event structs use these pairs to add abi tags, which keeps
			// decoding correct when a name is a Go keyword or normalizes to the same
			// identifier as another event field.
			inputs := make([]*tmplEventField, len(original.Inputs))
			for i := range original.Inputs {
				inputs[i] = &tmplEventField{Original: original.Inputs[i], Normalized: normalized.Inputs[i]}
			}
			// Append the event to the accumulator list. Anonymous events can still
			// be filtered by contract address and indexed topics; they simply omit
			// the event signature topic.
			events[original.Name] = &tmplEvent{Original: original, Normalized: normalized, Inputs: inputs}
		}
		// Add two special fallback functions if they exist
		if qrvmABI.HasFallback() {
			fallback = &tmplMethod{Original: qrvmABI.Fallback}
		}
		if qrvmABI.HasReceive() {
			receive = &tmplMethod{Original: qrvmABI.Receive}
		}
		contracts[types[i]] = &tmplContract{
			Type:        contractTypes[i],
			InputABI:    strippedABI,
			InputBin:    strings.TrimPrefix(strings.TrimSpace(bytecodes[i]), "0x"),
			Constructor: constructor,
			Calls:       calls,
			Transacts:   transacts,
			Fallback:    fallback,
			Receive:     receive,
			Events:      events,
			Libraries:   make(map[string]string),
		}
		// Function 4-byte signatures are stored in the same sequence
		// as types, if available.
		if len(fsigs) > i {
			contracts[types[i]].FuncSigs = fsigs[i]
		}
		// Parse library references.
		for pattern, name := range libs {
			matched, err := regexp.MatchString("__\\$"+pattern+"\\$__", contracts[types[i]].InputBin)
			if err != nil {
				log.Error("Could not search for pattern", "pattern", pattern, "contract", contracts[types[i]], "err", err)
			}
			if matched {
				contracts[types[i]].Libraries[pattern] = name
				// keep track that this type is a library
				if _, ok := isLib[name]; !ok {
					isLib[name] = struct{}{}
				}
			}
		}
	}
	// Check if that type has already been identified as a library
	for i := range types {
		_, ok := isLib[types[i]]
		contracts[types[i]].Library = ok
	}
	// Generate the contract template data content and render it
	data := &tmplData{
		Package:   pkg,
		Contracts: contracts,
		Libraries: libs,
		Structs:   structs,
	}
	buffer := new(bytes.Buffer)

	funcs := map[string]any{
		"bindtype":          bindType,
		"bindtopictype":     bindTopicType,
		"bindtopicruletype": bindTopicRuleType,
		"bindeventtag":      bindEventTag,
		"quote":             strconv.Quote,
		"capitalise":        abi.ToCamelCase,
		"decapitalise":      decapitalise,
	}
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplSource))
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}
	// Pass the code through gofmt to clean it up
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, buffer)
	}
	return string(code), nil
}

// bindBasicType converts basic hyperion types(except array, slice and tuple) to Go ones.
func bindBasicType(kind abi.Type) string {
	switch kind.T {
	case abi.AddressTy:
		return "common.Address"
	case abi.IntTy, abi.UintTy:
		parts := intRegex.FindStringSubmatch(kind.String())
		switch parts[2] {
		case "8", "16", "32", "64":
			return fmt.Sprintf("%sint%s", parts[1], parts[2])
		}
		return "*big.Int"
	case abi.FixedBytesTy:
		return fmt.Sprintf("[%d]byte", kind.Size)
	case abi.BytesTy:
		return "[]byte"
	case abi.FunctionTy:
		// Bind rejects FunctionTy before template execution. Keep this panic as
		// a defensive assertion so a new generation path cannot silently revive
		// the inherited 24-byte Ethereum representation on VM64.
		panic("bind: unsupported ABI function type reached code generation")
	default:
		// string, bool types
		return kind.String()
	}
}

// bindType converts hyperion types to Go ones. Since there is no clear mapping
// from all Hyperion types to Go ones (e.g. uint24), those that cannot be exactly
// mapped use *big.Int.
func bindType(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		return structs[bindStructTypeID(kind)].Name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindType(*kind.Elem, structs)
	case abi.SliceTy:
		return "[]" + bindType(*kind.Elem, structs)
	default:
		return bindBasicType(kind)
	}
}

// bindTopicType converts a Hyperion topic type to a Go one. Indexed values that
// cannot be reconstructed from a topic are exposed as their encoding hash.
func bindTopicType(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.StringTy, abi.BytesTy, abi.SliceTy, abi.ArrayTy, abi.TupleTy:
		return "common.Hash"
	default:
		return bindType(kind, structs)
	}
}

// bindTopicRuleType converts an indexed filter value to the type accepted by
// abi.MakeTopics. Strings and bytes remain preimages because MakeTopics hashes
// them. Arrays, slices, and tuples require a caller-supplied topic because
// their indexed encoding cannot be reconstructed by the generic ABI encoder,
// so callers provide the canonical Keccak-256 digest.
func bindTopicRuleType(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.SliceTy, abi.ArrayTy, abi.TupleTy:
		return "common.Hash"
	default:
		return bindType(kind, structs)
	}
}

// bindStructType converts a Hyperion tuple type to a Go one and records the mapping
// in the given map. Notably, this function will resolve and record nested struct
// recursively.
func bindStructType(kind abi.Type, structs map[string]*tmplStruct, packageNames map[string]bool) string {
	switch kind.T {
	case abi.TupleTy:
		id := bindStructTypeID(kind)
		if s, exist := structs[id]; exist {
			return s.Name
		}
		var (
			names  = make(map[string]bool)
			fields []*tmplField
		)
		for i, elem := range kind.TupleElems {
			name := abi.ToCamelCase(kind.TupleRawNames[i])
			name = abi.ResolveNameConflict(name, func(s string) bool { return names[s] })
			names[name] = true
			fields = append(fields, &tmplField{
				Type:         bindStructType(*elem, structs, packageNames),
				Name:         name,
				OriginalName: kind.TupleRawNames[i],
				SolKind:      *elem,
			})
		}
		name := kind.TupleRawName
		if name == "" {
			name = fmt.Sprintf("Struct%d", len(structs))
		}
		name = abi.ToCamelCase(name)
		if name == "" {
			name = fmt.Sprintf("Struct%d", len(structs))
		}
		name = abi.ResolveNameConflict(name, func(candidate string) bool { return packageNames[candidate] })
		packageNames[name] = true

		structs[id] = &tmplStruct{
			Name:   name,
			Fields: fields,
		}
		return name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindStructType(*kind.Elem, structs, packageNames)
	case abi.SliceTy:
		return "[]" + bindStructType(*kind.Elem, structs, packageNames)
	default:
		return bindBasicType(kind)
	}
}

// bindStructTypeID returns the identity used to deduplicate generated tuple
// declarations. The reflected structural type is always part of the identity:
// Solidity permits the same source name in different contracts, and removing
// dots from internalType can also collapse distinct names (A.B and AB). Reuse
// is safe only when both the normalized source identity and complete field
// structure match.
func bindStructTypeID(kind abi.Type) string {
	var identity strings.Builder
	appendBindTypeIdentity(&identity, kind)
	return identity.String()
}

// appendBindTypeIdentity records both ABI structure and original component
// names recursively. Reflected Go shapes alone are insufficient: names such as
// data and _data normalize to the same exported field but require distinct ABI
// tags in generated bindings.
func appendBindTypeIdentity(identity *strings.Builder, kind abi.Type) {
	switch kind.T {
	case abi.TupleTy:
		identity.WriteString("tuple(")
		identity.WriteString(strconv.Quote(kind.TupleRawName))
		for i, elem := range kind.TupleElems {
			identity.WriteByte(';')
			identity.WriteString(strconv.Quote(kind.TupleRawNames[i]))
			identity.WriteByte(':')
			appendBindTypeIdentity(identity, *elem)
		}
		identity.WriteByte(')')
	case abi.ArrayTy:
		fmt.Fprintf(identity, "array[%d](", kind.Size)
		appendBindTypeIdentity(identity, *kind.Elem)
		identity.WriteByte(')')
	case abi.SliceTy:
		identity.WriteString("slice(")
		appendBindTypeIdentity(identity, *kind.Elem)
		identity.WriteByte(')')
	default:
		identity.WriteString(strconv.Quote(kind.String()))
	}
}

// bindEventTag returns a quoted Go struct tag pairing a generated event field
// with its original ABI name. strconv.Quote keeps unusual, but validly parsed,
// ABI names from producing invalid Go source.
func bindEventTag(name string) string {
	return strconv.Quote("abi:" + strconv.Quote(name))
}

// rejectFunctionTypes prevents bindings from exposing the inherited Ethereum
// representation of a function value. On VM64 a function value combines a
// 64-byte address and a four-byte selector, so its 68 bytes cannot fit in one
// 64-byte ABI word. Function values need an explicit protocol representation
// before they can be safely generated.
func rejectFunctionTypes(contract string, parsed abi.ABI) error {
	unsupported := func(location string, arguments abi.Arguments) error {
		for _, argument := range arguments {
			if hasFunctionType(argument.Type) {
				return fmt.Errorf("binding %q: %s %q uses unsupported ABI function type: a VM64 function value is 68 bytes (64-byte address plus 4-byte selector) and cannot fit in one 64-byte ABI word", contract, location, argument.Name)
			}
		}
		return nil
	}
	if err := unsupported("constructor input", parsed.Constructor.Inputs); err != nil {
		return err
	}
	for _, method := range parsed.Methods {
		if err := unsupported(fmt.Sprintf("method %q input", method.RawName), method.Inputs); err != nil {
			return err
		}
		if err := unsupported(fmt.Sprintf("method %q output", method.RawName), method.Outputs); err != nil {
			return err
		}
	}
	for _, event := range parsed.Events {
		if err := unsupported(fmt.Sprintf("event %q input", event.RawName), event.Inputs); err != nil {
			return err
		}
	}
	for _, abiError := range parsed.Errors {
		if err := unsupported(fmt.Sprintf("error %q input", abiError.Name), abiError.Inputs); err != nil {
			return err
		}
	}
	return nil
}

// hasFunctionType finds function values nested under arrays, slices or tuples.
func hasFunctionType(kind abi.Type) bool {
	switch kind.T {
	case abi.FunctionTy:
		return true
	case abi.ArrayTy, abi.SliceTy:
		return hasFunctionType(*kind.Elem)
	case abi.TupleTy:
		for _, element := range kind.TupleElems {
			if hasFunctionType(*element) {
				return true
			}
		}
	}
	return false
}

// alias returns an alias of the given string based on the aliasing rules
// or returns itself if no rule is matched.
func alias(aliases map[string]string, n string) string {
	if alias, exist := aliases[n]; exist {
		return alias
	}
	return n
}

// decapitalise makes a camel-case string which starts with a lower case character.
func decapitalise(input string) string {
	if len(input) == 0 {
		return input
	}
	goForm := abi.ToCamelCase(input)
	return strings.ToLower(goForm[:1]) + goForm[1:]
}

// structured checks whether a list of ABI data types has enough information to
// operate through a proper Go struct or if flat returns are needed.
func structured(args abi.Arguments) bool {
	if len(args) < 2 {
		return false
	}
	exists := make(map[string]bool)
	for _, out := range args {
		// If the name is anonymous, we can't organize into a struct
		if out.Name == "" {
			return false
		}
		// If the field name is empty when normalized or collides (var, Var, _var, _Var),
		// we can't organize into a struct
		field := abi.ToCamelCase(out.Name)
		if field == "" || exists[field] {
			return false
		}
		exists[field] = true
	}
	return true
}

// hasStruct returns an indicator whether the given type is struct, struct slice
// or struct array.
func hasStruct(t abi.Type) bool {
	switch t.T {
	case abi.SliceTy:
		return hasStruct(*t.Elem)
	case abi.ArrayTy:
		return hasStruct(*t.Elem)
	case abi.TupleTy:
		return true
	default:
		return false
	}
}
