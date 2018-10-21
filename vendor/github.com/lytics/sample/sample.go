package sample

import (
	"fmt"
	"math/rand"
)

type Sampler struct {
	rnd *rand.Rand
}

// NewSampler instantiates a new Sampler, using the seed provided to create a new `rand.Rand` object.
func NewSampler(seed int64) *Sampler {
	return &Sampler{
		rand.New(rand.NewSource(seed)),
	}
}

type index []uint

func newIndex(n int) index {
	xindex := make(index, n)
	for i := 0; i < n; i++ {
		xindex[i] = uint(i)
	}
	return xindex
}

func (i index) Remove(which uint) index {
	if which >= uint(len(i)) {
		return i
	}

	if which == uint(len(i)-1) {
		return i[:which]
	}
	return append(i[:which], i[which+1:]...)
}

type vector []float64

func (v vector) Remove(which uint) vector {
	if which >= uint(len(v)) {
		return v
	}

	if which == uint(len(v)-1) {
		return v[:which]
	}

	return append(v[:which], v[which+1:]...)
}

func (v vector) Sum() float64 {
	var s float64
	for _, val := range v {
		s += val
	}
	return s
}

// scale scales each element by the sum of the vector
func (v vector) Scale() vector {
	scaled := make(vector, len(v))

	sum := v.Sum()
	for i, val := range v {
		scaled[i] = val / sum
	}

	return scaled
}

func (v vector) CumProb() vector {
	sum := v.Sum()
	cumprob := make([]float64, len(v))

	var cumsum float64
	for i, val := range v {
		cumprob[i] = cumsum
		cumsum += val / sum
	}

	return cumprob
}

func Find(w vector, val float64) int {
	i := find(w, val)
	if i < 0 {
		i = 0
	}
	return i
}

func find(w vector, val float64) int {
	for i, weight := range w {
		if weight > val {
			return i - 1
		}
	}

	return len(w) - 1
}

// SampleFloats takes a slice of floats and returns a sample of n
// elements with probability proportional to the weights. `Replace`
// specifies whether or not to sample with replacement.
func (s *Sampler) SampleFloats(x []float64, n int, replace bool, weights vector) ([]float64, error) {
	xindex := newIndex(len(x))
	index, err := sample(xindex, n, replace, weights, s.rnd)
	if err != nil {
		return nil, err
	}

	result := make([]float64, n)
	for i, val := range index {
		result[i] = x[val]
	}

	return result, nil
}

// SampleInts takes a slice of ints and returns a sample of n
// elements with probability proportional to the weights. `Replace`
// specifies whether or not to sample with replacement.
func (s *Sampler) SampleInts(x []int, n int, replace bool, weights vector) ([]int, error) {
	xindex := newIndex(len(x))
	index, err := sample(xindex, n, replace, weights, s.rnd)
	if err != nil {
		return nil, err
	}

	result := make([]int, n)
	for i, val := range index {
		result[i] = x[val]
	}

	return result, nil
}

// return the sample index to the other public functions
func sample(x index, n int, replace bool, weights vector, rnd *rand.Rand) ([]uint, error) {
	if weights != nil && len(x) != len(weights) {
		return nil, fmt.Errorf("length of x (%d) unequal to length of weights (%d)", len(x), len(weights))
	}

	if !replace && n > len(x) {
		return nil, fmt.Errorf("cannot sample with replacement when n (%d) is greater than x (%d)", n, len(x))
	}

	// cumulative probabilities
	var cumprob vector
	if weights != nil {
		cumprob = weights.CumProb()
	} else {
		cumprob = make(vector, len(x))
		nx := float64(len(x))
		for i, _ := range x {
			cumprob[i] = float64(i) / nx
		}
	}

	results := make([]uint, n)
	for i := 0; i < n; i++ {
		index := uint(Find(cumprob, rnd.Float64()))
		results[i] = x[index]

		// if sampling w/o replacement, remove the index and re-scale weights
		if !replace {
			x = x.Remove(index)
			cumprob = cumprob.Remove(index).Scale()
		}
	}

	return results, nil
}
