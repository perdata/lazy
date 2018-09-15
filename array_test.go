// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package lazy_test

import (
	"flag"
	"github.com/perdata/lazy"
	"math/rand"
	"strings"
	"testing"
)

func newArray(str string) lazy.Array {
	return lazy.Array{Count: len(str), Value: Slicer(str)}
}

func TestArray(t *testing.T) {
	zero := lazy.Array{}
	if zero.Count != 0 {
		t.Fatal("Zero initialization", zero)
	}

	if zz := zero.Slice(0, 0); toString(zz) != toString(zero) || zz.Count != 0 {
		t.Fatal("Zero slice", zz)
	}

	hello := zero.Splice(0, 0, newArray("hello"))
	if x := toString(hello); x != "hello" {
		t.Fatal("Initial Splicing fails", x)
	}

	if x := toString(hello.Slice(3, 0)); x != "" {
		t.Fatal("zero slice", x)
	}

	if x := toString(hello.Slice(0, 4)); x != "hell" {
		t.Fatal("Initial slice  fails", x, hello.Slice(0, 4))
	}

	jello := hello.Splice(0, 1, newArray("j"))
	if x := toString(jello); x != "jello" {
		t.Fatal("jello", x)
	}

	jimbo := jello.Splice(1, 3, newArray("imb"))
	if x := toString(jimbo); x != "jimbo" {
		t.Fatal("jimbo", x)
	}

	if x := toString(jimbo.Slice(2, 2)); x != "mb" {
		t.Fatal("mb", x)
	}

	jino := jimbo.Splice(2, 2, newArray("n"))
	if x := toString(jino); x != "jino" {
		t.Fatal("jino", x)
	}

	jinova := jino.Splice(4, 0, newArray("va"))
	if x := toString(jinova); x != "jinova" {
		t.Fatal("jinova", x)
	}

	djino := newArray("d").Splice(1, 0, jino)
	if x := toString(djino); x != "djino" {
		t.Fatal("djino", x)
	}

	djinojinova := djino.Splice(5, 0, jinova)
	if x := toString(djinojinova); x != "djinojinova" {
		t.Fatal("djinojinova", x)
	}
}

func TestInvalidOffsets(t *testing.T) {
	mustPanic := func(fn func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Failed to panic")
			}
		}()
		fn()
	}

	replace := newArray("replace")
	initial := newArray("hello")
	mustPanic(func() {
		initial.Slice(-1, 4)
	})
	mustPanic(func() {
		initial.Slice(1, -2)
	})
	mustPanic(func() {
		initial.Slice(3, 20)
	})
	mustPanic(func() {
		initial.Splice(-1, 4, replace)
	})
	mustPanic(func() {
		initial.Splice(1, -2, replace)
	})
	mustPanic(func() {
		initial.Splice(3, 20, replace)
	})
}

func TestRandomSplices(t *testing.T) {
	initRandomString(10)
	defer initRandomString(strlen)

	var a lazy.Array
	var s string

	limit := 10
	init := func(str string) {
		s = str
		a = lazy.Array{Limit: limit, Count: len(str), Value: Slicer(str)}
	}

	splice := func(offset, count int, r string) {
		s = s[:offset] + r + s[offset+count:]
		a = a.Splice(offset, count, lazy.Array{Limit: limit, Count: len(r), Value: Slicer(r)})
		if a.Limit <= 0 {
			s := Slicer("")
			a.ForEach(func(v interface{}, _ int) {
				s += v.(Slicer)
			})
			a = lazy.Array{Limit: limit, Count: len(string(s)), Value: s}
		}
		if toString(a) != s {
			t.Fatal("Splice diverged from string splice", s, toString(a.Slice(0, 6)), offset, count, r, "\n", a)
		}
	}

	benchmarkRun(500000, init, splice)
}

func BenchmarkLazy(b *testing.B) {
	var a lazy.Array
	limit := 10
	init := func(str string) {
		a = lazy.Array{Limit: limit, Count: len(str), Value: Slicer(str)}
	}
	splice := func(offset, count int, r string) {
		a = a.Splice(offset, count, lazy.Array{Limit: limit, Count: len(r), Value: Slicer(r)})
		if a.Limit <= 0 {
			var b strings.Builder
			b.Grow(a.Count)
			a.ForEach(func(v interface{}, _ int) {
				b.Write([]byte(string(v.(Slicer))))
			})
			a = lazy.Array{Limit: limit, Count: a.Count, Value: Slicer(b.String())}
		}
	}
	benchmark(b, init, splice)
}

func BenchmarkString(b *testing.B) {
	s := ""
	init := func(str string) {
		s = str
	}
	splice := func(offset, count int, r string) {
		s = s[:offset] + r + s[offset+count:]
	}
	benchmark(b, init, splice)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var largeRandomString string
var strlen int

func init() {
	strlen = 1000000
	flag.IntVar(&strlen, "strlen", 1000000, "length of large string to  use")
	flag.Parse()

	initRandomString(strlen)
}

func initRandomString(size int) {
	rand.Seed(42)
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	largeRandomString = string(b)
}

func benchmark(b *testing.B, init func(string), splice func(offset, count int, v string)) {
	for n := 0; n < b.N; n++ {
		benchmarkRun(100, init, splice)
	}
}

func benchmarkRun(iter int, init func(string), splice func(offset, count int, v string)) {
	rand.Seed(42)
	str := largeRandomString
	init(str)
	size := len(str)
	for kk := 0; kk < iter; kk++ {
		randAlpha := 'a' + rune(rand.Intn(26))
		v := string([]rune{randAlpha})
		offset, count := 0, 0

		if size > 0 {
			offset = rand.Intn(size)
		}

		diff := size - offset
		if diff > 100 {
			diff = 100
		}

		if diff > 0 {
			count = rand.Intn(diff)
		}

		splice(offset, count, v)
		size += 1 - count
	}
}

func toString(a lazy.Array) string {
	result := ""
	a.ForEach(func(leaf interface{}, count int) {
		result += string(leaf.(Slicer))
	})
	return result
}

type Slicer string

func (s Slicer) Slice(offset, count int) interface{} {
	return s[offset : offset+count]
}

func (s Slicer) Splice(offset, count int, replacement interface{}) interface{} {
	r := replacement.(Slicer)
	return s[:offset] + r + s[offset+count:]
}
