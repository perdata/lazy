// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// Package lazy implements a lazy collection.
//
// Slice and Splice simply defer the operation.  ForEach iterates
// through all the sub-slices and provides a way to reconstruct the
// new array.  This effectively prevents a lot of intermediate memory
// allocation and copying saving time (possibly at the cost of extra
// memory).
//
// Benchmark results
//
//    $ go test --bench=. --strlen=50000 -benchmem
//    goos: darwin
//    goarch: amd64
//    pkg: github.com/perdata/lazy
//    BenchmarkLazy-4     	    5000	    231841 ns/op	  920336 B/op	     722 allocs/op
//    BenchmarkString-4   	    2000	    501238 ns/op	 5046682 B/op	     200 allocs/op
//    PASS
//
package lazy

// Slicer is the interface that the base array should implement.
type Slicer interface {
	Slice(offset, count int) interface{}
}

// Array implements a lazy version of Slice and Splice. Each operation
// simply cause the args to stored without actually implementing
// it. The limit  is decremented with each deferred operation. This
// allows applications to wait till the limit is zeroed before calling
// ForEach and collecting all the segments into a simpler form.
//
// Creating a lazy array out of a non-lazy array:
//
//     lazyArray := lazy.Array{Limit:100, Count:count, Value:nonLazy}
//
// The limit can be omitted if it is not used  to decide when to
// flatten the array.
type Array struct {
	Limit, Count int
	Value        interface{}

	// offset is only set in case of a slice
	// rcount and rep are only set in case of a splice
	offset, rcount int
	replacement    *Array
}

// Slice simply stores the attempted slice without actually doing
// it. Use ForEach to visit all the segments of the underlying arrays
func (a Array) Slice(offset, count int) Array {
	if offset < 0 || count < 0 || offset+count > a.Count {
		panic("invalid slice args")
	}

	if offset == 0 && count == a.Count {
		return a
	}

	if count == 0 {
		return Array{Limit: a.Limit}
	}

	if a.replacement == nil {
		return Array{Limit: a.Limit, Count: count, Value: a.Value, offset: a.offset + offset}
	}

	return Array{Limit: a.Limit - 1, Count: count, Value: a, offset: offset}
}

// Splice simply stores the attempted splice. The operation is
// never executed but the underlying slices/segments can be iterated
// using ForEach
func (a Array) Splice(offset, count int, replacement Array) Array {
	if offset < 0 || count < 0 || offset+count > a.Count {
		panic("invalid splice args")
	}

	if offset == 0 && count == a.Count {
		return replacement
	}

	diff := replacement.Count - count
	value, limit := a.Value, a.Limit
	if a.offset != 0 || a.replacement != nil {
		value = a
		limit--
	}

	return Array{
		Limit:       limit,
		Count:       a.Count + diff,
		Value:       value,
		offset:      offset,
		rcount:      count,
		replacement: &replacement,
	}
}

// ForEach visits all the underlying segments.
func (a Array) ForEach(fn func(v interface{}, count int)) {
	a.forEach(0, a.Count, fn)
}

func forEach(v interface{}, offset, count int, fn func(interface{}, int)) {
	if count == 0 {
		return
	}

	if a, ok := v.(Array); ok {
		a.forEach(offset, count, fn)
	} else {
		fn(v.(Slicer).Slice(offset, count), count)
	}
}

func (a Array) forEach(offset, count int, fn func(interface{}, int)) {
	if count == 0 {
		return
	}

	if a.replacement != nil { // splice
		o, c := a.intersect(offset, count, 0, a.offset)
		forEach(a.Value, o, c, fn)
		o, c = a.intersect(offset, count, a.offset, a.replacement.Count)
		a.replacement.forEach(o-a.offset, c, fn)
		o, c = a.intersect(offset, count, a.offset+a.replacement.Count, a.Count)
		forEach(a.Value, o-a.replacement.Count+a.rcount, c, fn)
		return
	}

	forEach(a.Value, a.offset+offset, count, fn)
}

func (a Array) intersect(o1, c1, o2, c2 int) (int, int) {
	s := o1
	if o2 > s {
		s = o2
	}
	e := o1 + c1
	if o2+c2 < e {
		e = o2 + c2
	}
	if e > s {
		return s, e - s
	}
	return 0, 0
}
