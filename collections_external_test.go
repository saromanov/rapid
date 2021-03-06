// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package rapid_test

import (
	"reflect"
	"strconv"
	"testing"

	. "github.com/flyingmutant/rapid"
)

func TestSlicesOf(t *testing.T) {
	gens := []*Generator{
		SlicesOf(Booleans()),
		SlicesOf(Bytes()),
		SlicesOf(Ints()),
		SlicesOf(Uints()),
	}

	for _, g := range gens {
		t.Run(g.String(), MakeCheck(func(t *T, v interface{}) {
			if rv(v).Kind() != reflect.Slice {
				t.Fatalf("got not a slice")
			}

			Assume(rv(v).Len() > 0)
		}, g))
	}
}

func TestSlicesOfDistinct(t *testing.T) {
	g := SlicesOfDistinct(Ints(), nil)

	Check(t, func(t *T, s []int) {
		m := map[int]struct{}{}
		for _, i := range s {
			m[i] = struct{}{}
		}
		if len(m) != len(s) {
			t.Fatalf("%v unique out of %v", len(m), len(s))
		}
	}, g)
}

func TestSlicesOfDistinctBy(t *testing.T) {
	g := SlicesOfDistinct(Ints(), func(i int) string { return strconv.Itoa(i % 5) })

	Check(t, func(t *T, s []int) {
		m := map[int]struct{}{}
		for _, i := range s {
			m[i%5] = struct{}{}
		}
		if len(m) != len(s) {
			t.Fatalf("%v unique out of %v", len(m), len(s))
		}
	}, g)
}

func TestMapsOf(t *testing.T) {
	gens := []*Generator{
		MapsOf(Booleans(), Ints()),
		MapsOf(Ints(), Uints()),
		MapsOf(Uints(), SlicesOf(Booleans())),
	}

	for _, g := range gens {
		t.Run(g.String(), MakeCheck(func(t *T, v interface{}) {
			if rv(v).Kind() != reflect.Map {
				t.Fatalf("got not a map")
			}

			Assume(rv(v).Len() > 0)
		}, g))
	}
}

func TestMapsOfValues(t *testing.T) {
	g := MapsOfValues(Custom(genStruct), func(s testStruct) int { return s.x })

	Check(t, func(t *T, m map[int]testStruct) {
		for k, v := range m {
			if k != v.x {
				t.Fatalf("got key %v with value %v", k, v)
			}
		}
	}, g)
}

func TestArraysOf(t *testing.T) {
	elems := []*Generator{Booleans(), Ints(), Uints()}
	counts := []int{0, 1, 3, 17}

	for _, e := range elems {
		for _, c := range counts {
			g := ArraysOf(c, e)
			t.Run(g.String(), MakeCheck(func(t *T, v interface{}) {
				if rv(v).Len() != c {
					t.Fatalf("len is %v instead of %v", rv(v).Len(), c)
				}
			}, g))
		}
	}
}

func TestCollectionLenLimits(t *testing.T) {
	genFuncs := []func(i, j int) *Generator{
		func(i, j int) *Generator { return StringsOfNBytes(i, j) },
		func(i, j int) *Generator { return SlicesOfN(Bytes(), i, j) },
		func(i, j int) *Generator { return SlicesOfNDistinct(Bytes(), i, j, nil) },
		func(i, j int) *Generator { return SlicesOfNDistinct(Ints(), i, j, func(n int) int { return n % j }) },
		func(i, j int) *Generator { return MapsOfN(Ints(), Ints(), i, j) },
		func(i, j int) *Generator { return MapsOfNValues(Ints(), i, j, nil) },
		func(i, j int) *Generator { return MapsOfNValues(Ints(), i, j, func(n int) int { return n % j }) },
	}

	for i, gf := range genFuncs {
		t.Run(strconv.Itoa(i), MakeCheck(func(t *T, minLen int, maxLen int) {
			Assume(minLen <= maxLen)
			s := rv(t.Draw(gf(minLen, maxLen), "s"))
			if s.Len() < minLen {
				t.Fatalf("got collection of length %v with minLen %v", s.Len(), minLen)
			}
			if s.Len() > maxLen {
				t.Fatalf("got collection of length %v with maxLen %v", s.Len(), maxLen)
			}
		}, IntsRange(0, 256), IntsMin(0)))
	}
}
