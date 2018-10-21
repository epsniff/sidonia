sample
======

The `sample` package implements weighted sampling in [Go](http://golang.org).

[![Build Status](https://travis-ci.org/lytics/sample.svg?branch=master)](https://travis-ci.org/lytics/sample) [![GoDoc](https://godoc.org/github.com/lytics/sample?status.svg)](https://godoc.org/github.com/lytics/sample)

## Usage

```go

// initialize a new sampler
sampler := sample.NewSampler(time.Now().UnixNano())

// sample proportionately to the weights provided
// note that the weights don't have to sum to 1
x := []int{1, 3, 2, 5, 3, 2}
weights := []float{0.6, 3.2, 2.1, 0.05, 1.0, 1.5}

n := 3          // sample size
replace := true // sample with replacement

result := sampler.SampleInts(x, n, replace, weights)
```

Or, if sampling without weights (for uniform sampling)

```
result := sampler.SampleInts(x, n, replace, nil)
```
