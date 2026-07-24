// Copyright 2022 The go-ethereum Authors
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

package abi

import (
	"errors"
	"fmt"
)

type SelectorMarshaling struct {
	Name   string               `json:"name"`
	Type   string               `json:"type"`
	Inputs []ArgumentMarshaling `json:"inputs"`
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentifierSymbol(c byte) bool {
	return c == '$' || c == '_'
}

func parseToken(unescapedSelector string, isIdent bool) (string, string, error) {
	if len(unescapedSelector) == 0 {
		return "", "", errors.New("empty token")
	}
	firstChar := unescapedSelector[0]
	position := 1
	if !(isAlpha(firstChar) || (isIdent && isIdentifierSymbol(firstChar))) {
		return "", "", fmt.Errorf("invalid token start: %c", firstChar)
	}
	for position < len(unescapedSelector) {
		char := unescapedSelector[position]
		if !(isAlpha(char) || isDigit(char) || (isIdent && isIdentifierSymbol(char))) {
			break
		}
		position++
	}
	return unescapedSelector[:position], unescapedSelector[position:], nil
}

func parseIdentifier(unescapedSelector string) (string, string, error) {
	return parseToken(unescapedSelector, true)
}

func parseArraySuffix(unescapedSelector string) (string, string, error) {
	position := 0
	for position < len(unescapedSelector) && unescapedSelector[position] == '[' {
		position++
		for position < len(unescapedSelector) && isDigit(unescapedSelector[position]) {
			position++
		}
		if position == len(unescapedSelector) {
			return "", "", errors.New("failed to parse array: expected ']', got end of selector")
		}
		if unescapedSelector[position] != ']' {
			return "", "", fmt.Errorf("failed to parse array: expected ']', got %c", unescapedSelector[position])
		}
		position++
	}
	return unescapedSelector[:position], unescapedSelector[position:], nil
}

func parseElementaryType(unescapedSelector string) (string, string, error) {
	parsedType, rest, err := parseToken(unescapedSelector, false)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse elementary type: %v", err)
	}
	if parsedType == "tuple" {
		return "", "", errors.New("tuple type must use parenthesized components")
	}
	arraySuffix, rest, err := parseArraySuffix(rest)
	if err != nil {
		return "", "", err
	}
	return parsedType + arraySuffix, rest, nil
}

type parsedCompositeType struct {
	components  []any
	arraySuffix string
}

// parseCompositeType parses a parenthesized list whose members start at depth.
// The selector's outer argument list is not itself an ABI tuple, so its members
// start at depth zero. Tuple members advance the same nesting counter used by
// NewType, preventing hostile selectors from recursing before NewType gets a
// chance to enforce its safety limit.
func parseCompositeType(unescapedSelector string, depth int) (parsedCompositeType, string, error) {
	if len(unescapedSelector) == 0 {
		return parsedCompositeType{}, "", errors.New("expected '(', got end of selector")
	}
	if unescapedSelector[0] != '(' {
		return parsedCompositeType{}, "", fmt.Errorf("expected '(', got %c", unescapedSelector[0])
	}

	result := make([]any, 0)
	rest := unescapedSelector[1:]
	if len(rest) == 0 {
		return parsedCompositeType{}, "", errors.New("expected type or ')', got end of selector")
	}
	if rest[0] != ')' {
		for {
			parsedType, remaining, err := parseType(rest, depth)
			if err != nil {
				return parsedCompositeType{}, "", fmt.Errorf("failed to parse type: %v", err)
			}
			result = append(result, parsedType)
			rest = remaining
			if len(rest) == 0 {
				return parsedCompositeType{}, "", errors.New("expected ',' or ')', got end of selector")
			}
			if rest[0] == ')' {
				break
			}
			if rest[0] != ',' {
				return parsedCompositeType{}, "", fmt.Errorf("expected ',' or ')', got %c", rest[0])
			}
			rest = rest[1:]
			if len(rest) == 0 {
				return parsedCompositeType{}, "", errors.New("expected type after ',', got end of selector")
			}
		}
	}

	// Consume the tuple's closing parenthesis, then retain every array
	// dimension. Tuple arrays use the exact same suffix grammar as elementary
	// arrays, including mixed fixed and dynamic dimensions.
	rest = rest[1:]
	arraySuffix, rest, err := parseArraySuffix(rest)
	if err != nil {
		return parsedCompositeType{}, "", err
	}
	return parsedCompositeType{components: result, arraySuffix: arraySuffix}, rest, nil
}

func parseType(unescapedSelector string, depth int) (any, string, error) {
	if len(unescapedSelector) == 0 {
		return nil, "", errors.New("empty type")
	}
	if depth > maxABITypeNesting {
		return nil, "", fmt.Errorf("abi: type nesting exceeds safety limit %d", maxABITypeNesting)
	}
	if unescapedSelector[0] == '(' {
		return parseCompositeType(unescapedSelector, depth+1)
	} else {
		return parseElementaryType(unescapedSelector)
	}
}

func assembleArgs(args []any) ([]ArgumentMarshaling, error) {
	arguments := make([]ArgumentMarshaling, 0)
	for i, arg := range args {
		// generate dummy name to avoid unmarshal issues
		name := fmt.Sprintf("name%d", i)
		if s, ok := arg.(string); ok {
			arguments = append(arguments, ArgumentMarshaling{name, s, s, nil, false})
		} else if composite, ok := arg.(parsedCompositeType); ok {
			subArgs, err := assembleArgs(composite.components)
			if err != nil {
				return nil, fmt.Errorf("failed to assemble components: %v", err)
			}
			tupleType := "tuple" + composite.arraySuffix
			arguments = append(arguments, ArgumentMarshaling{name, tupleType, tupleType, subArgs, false})
		} else {
			return nil, fmt.Errorf("failed to assemble args: unexpected type %T", arg)
		}
	}
	return arguments, nil
}

// ParseSelector converts a method selector into a struct that can be JSON encoded
// and consumed by other functions in this package.
// Note, although uppercase letters are not part of the ABI spec, this function
// still accepts it as the general format is valid.
func ParseSelector(unescapedSelector string) (SelectorMarshaling, error) {
	name, rest, err := parseIdentifier(unescapedSelector)
	if err != nil {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': %v", unescapedSelector, err)
	}
	parsedArgs, rest, err := parseCompositeType(rest, 0)
	if err != nil {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': %v", unescapedSelector, err)
	}
	if parsedArgs.arraySuffix != "" {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': function argument list cannot have array suffix %q", unescapedSelector, parsedArgs.arraySuffix)
	}
	if len(rest) > 0 {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': unexpected string '%s'", unescapedSelector, rest)
	}

	// Reassemble the fake ABI and construct the JSON
	fakeArgs, err := assembleArgs(parsedArgs.components)
	if err != nil {
		return SelectorMarshaling{}, fmt.Errorf("failed to parse selector: %v", err)
	}
	for _, arg := range fakeArgs {
		if _, err := NewType(arg.Type, arg.InternalType, arg.Components); err != nil {
			return SelectorMarshaling{}, fmt.Errorf("failed to parse selector '%s': invalid argument type %q: %v", unescapedSelector, arg.Type, err)
		}
	}

	return SelectorMarshaling{name, "function", fakeArgs}, nil
}
