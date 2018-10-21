# govector

[![Build Status](https://travis-ci.org/drewlanenga/govector.svg?branch=master)](https://travis-ci.org/drewlanenga/govector) [![GoDoc](https://godoc.org/github.com/drewlanenga/govector?status.svg)](https://godoc.org/github.com/drewlanenga/govector)


Provide nice vector API for handling numeric types in [Go](http://golang.org).

## Usage

```go
// create a Vector type from an int array
x, err := AsVector([]int{1, 2, 3, 4, 6, 5})

// create a Vector type from a float64 array, to be used for weights
w, _ := AsVector([]float64{1.0, 1.0, 1.0, 1.0, 1.0, 4.0})

// find the differences of the Vector x
d_x := x.Diff()

// Generate the empirical CDF function for x
empirical := x.Ecdf()

// Calculate the percentile from the empirical CDF of x
percentile = empirical(2.4)

// Calculate the weighted mean of x using weights w
wm, _ = x.WeightedMean(w)

// Calculate the 5% and 95% quantiles of x
q, _ := AsVector([]float64{0.05, 0.95})
quantiles, _ = x.Quantiles(q)

// cumulative sum
s = x.Cumsum()

// shuffle x
shuffled := x.Shuffle()

// Apply arbitrary functions to vectors
_ = x.Apply(empirical)
_ = x.Apply(math.Sqrt)
```
