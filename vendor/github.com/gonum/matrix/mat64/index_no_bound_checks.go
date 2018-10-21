// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file must be kept in sync with index_bound_checks.go.

//+build !bounds

package mat64

import "github.com/gonum/matrix"

// At returns the element at row r, column c.
func (m *Dense) At(r, c int) float64 {
	if r >= m.mat.Rows || r < 0 {
		panic(matrix.ErrRowAccess)
	}
	if c >= m.mat.Cols || c < 0 {
		panic(matrix.ErrColAccess)
	}
	return m.at(r, c)
}

func (m *Dense) at(r, c int) float64 {
	return m.mat.Data[r*m.mat.Stride+c]
}

// Set sets the element at row r, column c to the value v.
func (m *Dense) Set(r, c int, v float64) {
	if r >= m.mat.Rows || r < 0 {
		panic(matrix.ErrRowAccess)
	}
	if c >= m.mat.Cols || c < 0 {
		panic(matrix.ErrColAccess)
	}
	m.set(r, c, v)
}

func (m *Dense) set(r, c int, v float64) {
	m.mat.Data[r*m.mat.Stride+c] = v
}

// At returns the element at row r, column c. It panics if c is not zero.
func (v *Vector) At(r, c int) float64 {
	if r < 0 || r >= v.n {
		panic(matrix.ErrRowAccess)
	}
	if c != 0 {
		panic(matrix.ErrColAccess)
	}
	return v.at(r)
}

func (v *Vector) at(r int) float64 {
	return v.mat.Data[r*v.mat.Inc]
}

// Set sets the element at row r to the value val. It panics if r is less than
// zero or greater than the length.
func (v *Vector) SetVec(i int, val float64) {
	if i < 0 || i >= v.n {
		panic(matrix.ErrVectorAccess)
	}
	v.setVec(i, val)
}

func (v *Vector) setVec(i int, val float64) {
	v.mat.Data[i*v.mat.Inc] = val
}

// At returns the element at row r and column c.
func (s *SymDense) At(r, c int) float64 {
	if r >= s.mat.N || r < 0 {
		panic(matrix.ErrRowAccess)
	}
	if c >= s.mat.N || c < 0 {
		panic(matrix.ErrColAccess)
	}
	return s.at(r, c)
}

func (s *SymDense) at(r, c int) float64 {
	if r > c {
		r, c = c, r
	}
	return s.mat.Data[r*s.mat.Stride+c]
}

// SetSym sets the elements at (r,c) and (c,r) to the value v.
func (s *SymDense) SetSym(r, c int, v float64) {
	if r >= s.mat.N || r < 0 {
		panic(matrix.ErrRowAccess)
	}
	if c >= s.mat.N || c < 0 {
		panic(matrix.ErrColAccess)
	}
	s.set(r, c, v)
}

func (s *SymDense) set(r, c int, v float64) {
	if r > c {
		r, c = c, r
	}
	s.mat.Data[r*s.mat.Stride+c] = v
}

// At returns the element at row r, column c.
func (t *TriDense) At(r, c int) float64 {
	if r >= t.mat.N || r < 0 {
		panic(matrix.ErrRowAccess)
	}
	if c >= t.mat.N || c < 0 {
		panic(matrix.ErrColAccess)
	}
	return t.at(r, c)
}

func (t *TriDense) at(r, c int) float64 {
	isUpper := t.isUpper()
	if (isUpper && r > c) || (!isUpper && r < c) {
		return 0
	}
	return t.mat.Data[r*t.mat.Stride+c]
}

// SetTri sets the element at row r, column c to the value v.
// It panics if the location is outside the appropriate half of the matrix.
func (t *TriDense) SetTri(r, c int, v float64) {
	if r >= t.mat.N || r < 0 {
		panic(matrix.ErrRowAccess)
	}
	if c >= t.mat.N || c < 0 {
		panic(matrix.ErrColAccess)
	}
	isUpper := t.isUpper()
	if (isUpper && r > c) || (!isUpper && r < c) {
		panic(matrix.ErrTriangleSet)
	}
	t.set(r, c, v)
}

func (t *TriDense) set(r, c int, v float64) {
	t.mat.Data[r*t.mat.Stride+c] = v
}
