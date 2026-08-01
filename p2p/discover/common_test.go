// Copyright 2026 The go-ethereum Authors
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

package discover

import (
	"container/list"
	"slices"
	"testing"
)

func TestIterListRemove(t *testing.T) {
	items := list.New()
	for i := range 4 {
		items.PushBack(i)
	}

	var visited []int
	for value, element := range iterList[int](items) {
		visited = append(visited, value)
		items.Remove(element)
	}
	if want := []int{0, 1, 2, 3}; !slices.Equal(visited, want) {
		t.Fatalf("visited elements mismatch: have %v, want %v", visited, want)
	}
	if items.Len() != 0 {
		t.Fatalf("list still contains %d elements", items.Len())
	}
}
