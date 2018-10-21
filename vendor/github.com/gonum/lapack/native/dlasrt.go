// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import (
	"sort"

	"github.com/gonum/lapack"
)

// Dlasrt sorts the numbers in the input slice d. If sort == lapack.SortIncreasing,
// the elements are sorted in increasing order. If sort == lapack.SortDecreasing,
// the elements are sorted in decreasing order.
func (impl Implementation) Dlasrt(s lapack.Sort, n int, d []float64) {
	checkVector(n, d, 1)
	d = d[:n]
	switch s {
	default:
		panic("lapack: bad sort")
	case lapack.SortIncreasing:
		sort.Sort(sort.Reverse(sort.Float64Slice(d)))
	case lapack.SortDecreasing:
		sort.Float64s(d)
	}
}
