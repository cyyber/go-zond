// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package abi

import (
	"math/big"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/uint512"
)

// validIntegerSize bounds the declared width to one ABI word. Widths that are
// not multiples of 8 remain accepted for compatibility with the inherited ABI
// type parser.
func validIntegerSize(size int) bool {
	return size >= 1 && size <= uint512.WordBits
}

func maxUnsignedInteger(size int) *big.Int {
	switch size {
	case 256:
		return new(big.Int).Set(MaxUint256)
	case uint512.WordBits:
		return new(big.Int).Set(MaxUint512)
	default:
		return new(big.Int).Sub(new(big.Int).Lsh(common.Big1, uint(size)), common.Big1)
	}
}

func maxSignedInteger(size int) *big.Int {
	switch size {
	case 256:
		return new(big.Int).Set(MaxInt256)
	default:
		return new(big.Int).Sub(new(big.Int).Lsh(common.Big1, uint(size-1)), common.Big1)
	}
}

func minSignedInteger(size int) *big.Int {
	return new(big.Int).Neg(new(big.Int).Lsh(common.Big1, uint(size-1)))
}

func fitsUnsignedInteger(value *big.Int, size int) bool {
	return validIntegerSize(size) && value.Sign() >= 0 && value.Cmp(maxUnsignedInteger(size)) <= 0
}

func fitsSignedInteger(value *big.Int, size int) bool {
	if !validIntegerSize(size) {
		return false
	}
	return value.Cmp(minSignedInteger(size)) >= 0 && value.Cmp(maxSignedInteger(size)) <= 0
}
