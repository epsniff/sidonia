package govector

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const (
	NA = math.SmallestNonzeroFloat64
)

// rnd is a private prng so we don't alter global prng state
var (
	rnd      = rand.New(rand.NewSource(time.Now().UnixNano()))
	rndMutex = &sync.Mutex{}
)

type Vector []float64

// Copy returns a copy the input vector.  This is useful for functions that
// perform modification and shuffling on the order of the input vector.
func (x Vector) Copy() Vector {
	y := make(Vector, len(x))
	copy(y, x)
	return y
}

// Smooth takes a sliding window average of vector. Indices i and j refer to the
// the number of points you'd like to consider before and after a point in
// the average.
func (x Vector) Smooth(left, right uint) Vector {
	n := uint(len(x))
	smoothed := make(Vector, n)

	for index := uint(0); index < n; index++ {
		var leftmost uint
		if left < index {
			leftmost = index - left
		}

		rightmost := index + right + 1
		if rightmost > n {
			rightmost = n
		}

		window := x[leftmost:rightmost]
		smoothed[index] = window.Mean()
	}

	return smoothed
}

// Len, Swap, and Less are implemented to allow for direct
// sorting on Vector types.
func (x Vector) Len() int {
	return len(x)
}

func (x Vector) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x Vector) Less(i, j int) bool {
	return x[i] < x[j]
}

func (x Vector) Sort() {
	sort.Sort(x)
}

// Sum returns the sum of the vector.
func (x Vector) Sum() float64 {
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s
}

// Abs returns the absolute values of the vector elements.
func (x Vector) Abs() Vector {
	y := x.Copy()

	for i, _ := range y {
		y[i] = math.Abs(y[i])
	}

	return y
}

// Cumsum returns the cumulative sum of the vector.
func (x Vector) Cumsum() Vector {
	y := make(Vector, len(x))

	y[0] = x[0]

	i := 1
	for i < len(x) {
		y[i] = x[i] + y[i-1]
		i++
	}

	return y
}

// Mean returns the mean of the vector.
func (x Vector) Mean() float64 {
	s := x.Sum()

	n := float64(len(x))

	return s / n
}

// weightedSum returns the weighted sum of the vector.  This is really only useful in
// calculating the weighted mean.
func (x Vector) weightedSum(w Vector) (float64, error) {
	if len(x) != len(w) {
		return NA, fmt.Errorf("Length of weights unequal to vector length")
	}

	ws := 0.0
	for i, _ := range x {
		ws += x[i] * w[i]
	}
	return ws, nil
}

// WeightedMean returns the weighted mean of the vector for a given vector of weights.
func (x Vector) WeightedMean(w Vector) (float64, error) {
	ws, err := x.weightedSum(w)
	if err != nil {
		return NA, err
	}
	sw := w.Sum()

	return ws / sw, nil
}

func (x Vector) variance(mean float64) float64 {
	n := float64(len(x))
	if n == 1 {
		return 0
	} else if n < 2 {
		n = 2
	}

	ss := 0.0
	for _, v := range x {
		ss += math.Pow(v-mean, 2.0)
	}

	return ss / (n - 1)
}

// Variance caclulates the variance of the vector
func (x Vector) Variance() float64 {
	return x.variance(x.Mean())
}

// MeanVar returns both the mean and the variance of the vector
func (x Vector) MeanVar() (float64, float64) {
	m := x.Mean()
	v := x.variance(m)
	return m, v
}

// Sd calculates the standard deviation of the vector
func (x Vector) Sd() float64 {
	return math.Sqrt(x.Variance())
}

// Max returns the maximum value of the vector
func (x Vector) Max() float64 {
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}
	return max
}

// Min returns the minimum value of the vector
func (x Vector) Min() float64 {
	min := x[0]
	for _, v := range x {
		if v < min {
			min = v
		}
	}
	return min
}

// Ecdf returns the empirical cumulative distribution function.  The ECDF function
// will return the percentile of a given value relative to the vector.
func (x Vector) Ecdf() func(float64) float64 {
	y := x.Copy()

	y.Sort()
	n := len(y)

	empirical := func(q float64) float64 {
		i := 0
		for i < n {
			if q < y[i] {
				return float64(i) / float64(n)
			}
			i++
		}
		return 1.0
	}

	return empirical
}

// Apply returns the values of the vector applied to an arbitrary function, which must
// return a float64, since a Vector will be returned.
func (x Vector) Apply(f func(float64) float64) Vector {
	y := make(Vector, len(x))

	for i, v := range x {
		y[i] = f(v)
	}
	return y
}

// Filter returns the values that match the filter function.  Vector elements with return
// values of TRUE are filtered/removed.
func (x Vector) Filter(f func(float64) bool) Vector {
	y := make(Vector, 0, len(x))

	for _, v := range x {
		if !f(v) {
			y = append(y, v)
		}
	}

	return y
}

// Quantiles returns the quantiles of a vector corresponding to input quantiles using a
// weighted average approach for index interpolation.
func (x Vector) Quantiles(q Vector) Vector {
	y := x.Copy()

	y.Sort()

	n := float64(len(y))
	output := make(Vector, len(q))
	for i, quantile := range q {

		if n == 0.0 {
			output[i] = 0
			continue
		}

		fuzzyQuantile := quantile * n

		// the quantile lies directly on the value
		if fuzzyQuantile-math.Floor(fuzzyQuantile) == 0.5 {
			output[i] = float64(y[int(math.Floor(fuzzyQuantile))])
			continue
		}

		lowerIndex := math.Max(0, math.Floor(fuzzyQuantile)-1)
		upperIndex := math.Min(lowerIndex+1, n-1)

		values := Vector{float64(y[int(lowerIndex)]), float64(y[int(upperIndex)])}

		indexDiff := fuzzyQuantile - math.Floor(fuzzyQuantile)

		lowerWeight := 1.0
		upperWeight := 1.0

		if indexDiff > 0.0 {
			lowerWeight = 1.0 - indexDiff
			upperWeight = indexDiff
		}

		output[i], _ = values.WeightedMean(Vector{lowerWeight, upperWeight})
	}

	return output
}

// Diff returns a vector of length (n - 1) of the differences in the input vector
func (x Vector) Diff() Vector {
	n := len(x)

	if n < 2 {
		return Vector{NA}
	} else {
		d := make(Vector, n-1)

		i := 1
		for i < n {
			d[i-1] = x[i] - x[i-1]
			i++
		}
		return d
	}
}

// RelDiff returns a vector of the relative differences of the input vector
func (x Vector) RelDiff() Vector {
	n := len(x)

	if n < 2 {
		return Vector{NA}
	} else {
		d := make(Vector, n-1)

		i := 1
		for i < n {
			d[i-1] = (x[i] - x[i-1]) / x[i]
			i++
		}
		return d
	}
}

// Sample returns a sample of n elements of the original input vector.
func (x Vector) Sample(n int) Vector {
	// unprotected access to custom rand.Rand objects can cause panics
	// https://github.com/golang/go/issues/3611
	rndMutex.Lock()
	perm := rnd.Perm(len(x))
	rndMutex.Unlock()

	// sample n elements
	perm = perm[:n]

	y := make(Vector, n)
	for yi, permi := range perm {
		y[yi] = x[permi]
	}

	return y
}

// Shuffle returns a shuffled copy of the original input vector.
func (x Vector) Shuffle() Vector {
	return x.Sample(len(x))
}

// Join returns an (efficiently joined) vector of the input vectors.
func Join(vectors ...Vector) Vector {
	// figure out how big to make the resulting vector so we can
	// allocate efficiently
	n := 0
	for _, vector := range vectors {
		n += vector.Len()
	}

	i := 0
	v := make(Vector, n)
	for _, vector := range vectors {
		for _, value := range vector {
			v[i] = value
			i++
		}
	}

	return v
}

// Rank returns a vector of the ranked values of the input vector.
func (x Vector) Rank() Vector {
	y := x.Copy()
	y.Sort()

	// equivalent to a minimum rank (tie) method
	rank := 0
	ranks := make(Vector, len(x))
	for i, _ := range ranks {
		ranks[i] = -1
	}

	for i, _ := range y {
		for j, _ := range x {
			if y[i] == x[j] && ranks[j] == -1 {
				ranks[j] = float64(rank)
			}
		}
		rank++
	}

	return ranks
}

// Order returns a vector of untied ranks of the input vector.
func (x Vector) Order() Vector {
	y := x.Copy()
	y.Sort()

	rank := 0
	order := make(Vector, len(x))
	for i, _ := range order {
		order[i] = -1
	}
	for i, _ := range y {
		for j, _ := range x {
			if y[i] == x[j] && order[j] == -1 {
				order[j] = float64(rank)
				rank++
				break
			}
		}
	}
	return order
}

// Push appends the input vector with the value to be pushed.
func (x *Vector) Push(y float64) {
	*x = append(*x, y)
	return
}

//Append values to an array. Array size will not grow if unnecessary.
//It will grow if the cap has been extended by external modification.
func (x *Vector) PushFixed(y float64) error {
	lenx := len(*x)
	if lenx <= cap(*x) {
		slicex := (*x)[1:]
		z := make([]float64, lenx, lenx)
		copy(z, slicex)
		z[lenx-1] = y
		*x = z
		return nil
	} else {
		return fmt.Errorf("GoVector length greater than capacity!? len: %d cap: %d\n%#v", len(*x), cap(*x), x)
	}
}
