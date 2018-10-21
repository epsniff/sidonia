package govector

import (
	"fmt"
	"math"
)

// Product returns a vector of element-wise products of two input vectors.
func Product(x, y Vector) (Vector, error) {
	if len(x) != len(y) {
		return nil, fmt.Errorf("x and y have unequal lengths: %d / %d", len(x), len(y))
	}

	p := make(Vector, len(x))
	for i, _ := range x {
		p[i] = x[i] * y[i]
	}
	return p, nil
}

// DotProduct returns the dot product of two vectors.
func DotProduct(x, y Vector) (float64, error) {
	p, err := Product(x, y)
	if err != nil {
		return NA, err
	}
	return p.Sum(), nil
}

// Norm returns the vector norm.  Use pow = 2.0 for Euclidean.
func Norm(x Vector, pow float64) float64 {
	s := 0.0

	for _, xval := range x {
		s += math.Pow(xval, pow)
	}

	return math.Pow(s, 1/pow)
}

// Cosine returns the cosine similarity between two vectors.
func Cosine(x, y Vector) (float64, error) {
	d, err := DotProduct(x, y)
	if err != nil {
		return NA, err
	}

	xnorm := Norm(x, 2.0)
	ynorm := Norm(y, 2.0)

	return d / (xnorm * ynorm), nil
}

// Cor returns the Pearson correlation between two vectors.
func Cor(x, y Vector) (float64, error) {
	n := float64(len(x))
	xy, err := Product(x, y)
	if err != nil {
		return NA, err
	}

	sx := x.Sd()
	sy := y.Sd()

	mx := x.Mean()
	my := y.Mean()

	r := (xy.Sum() - n*mx*my) / ((n - 1) * sx * sy)
	return r, nil
}
