# lazy

[![Status](https://travis-ci.com/perdata/lazy.svg?branch=master)](https://travis-ci.com/perdata/lazy?branch=master)
[![GoDoc](https://godoc.org/github.com/perdata/lazy?status.svg)](https://godoc.org/github.com/perdata/lazy)
[![codecov](https://codecov.io/gh/perdata/lazy/branch/master/graph/badge.svg)](https://codecov.io/gh/perdata/lazy)
[![GoReportCard](https://goreportcard.com/badge/github.com/perdata/lazy)](https://goreportcard.com/report/github.com/perdata/lazy)

Lazy is a golang package for dealing with mutations on large
arrays. It works by basically deferring the actual edit methods and
providing a way to iterate through all the slices. This avoids
large-scale memory copying.

This still involves a fair degree of copying and very large arrays
will benefit from the [lazy package](https://github.com/perdata/trope) instead.

## Benchmark stats


```sh
$ go test --bench=. --strlen=50000 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/perdata/lazy
BenchmarkLazy-4     	    5000	    231841 ns/op	  920336 B/op	     722 allocs/op
BenchmarkString-4   	    2000	    501238 ns/op	 5046682 B/op	     200 allocs/op
PASS
````

