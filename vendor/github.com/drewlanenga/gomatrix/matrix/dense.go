// Copyright 2009 The GoMatrix Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package matrix implements a simple library for
// matrix operations and linear algebra.
package matrix

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
)

// A matrix backed by a flat array of all elements.
type DenseMatrix struct {
	matrix
	// flattened matrix data. elements[i*step+j] is row i, col j
	elements []float64
	// actual offset between rows
	step int
}

// Pow raises every element of the matrix to power.  Returns a new
// matrix
func (A *DenseMatrix) Pow(power float64) *DenseMatrix {
	numRows, numCols := A.GetSize()
	raised := Zeros(numRows, numCols)

	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			raised.Set(i, j, math.Pow(A.Get(i, j), power))
		}
	}
	return raised
}

// Squares every element of the matrix.  Returns a new
// matrix.
func (A DenseMatrix) Sqr() *DenseMatrix {
	numRows, numCols := A.GetSize()
	squared := Zeros(numRows, numCols)

	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			x := A.Get(i, j)
			squared.Set(i, j, x*x)
		}
	}
	return squared
}

// Sqrm squares every value in the matrix. In modifies the matix in place.
func (A *DenseMatrix) Sqrm() {
	numRows, numCols := A.GetSize()
	//	squared := Zeros(numRows, numCols)

	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			x := A.Get(i, j)
			A.Set(i, j, x*x)
		}
	}
}

// FiltCol find values that match min <= A <= max for a specific column.
//
// Returns the rows that match.
func (A *DenseMatrix) FiltCol(min, max float64, col int) (*DenseMatrix, error) {
	rows, cols := A.GetSize()
	buf := make([]float64, 0)

	if col < 0 || col > cols-1 {
		matches := Zeros(1, 1)
		return matches, errors.New(fmt.Sprintf("FiltCol: Expected col vaule in range 0 to %d.  Received %d\n", cols-1, col))
	}

	num_matches := 0
	for i := 0; i < rows; i++ {
		v := A.Get(i, col)

		if v >= min && v <= max {
			for k := 0; k < cols; k++ {
				buf = append(buf, A.Get(i, k))
			}
			num_matches++
		}
	}

	matches := MakeDenseMatrix(buf, len(buf)/cols, cols)
	var err error = nil
	if num_matches == 0 {
		err = errors.New("matutil: no match")
	}
	return matches, err
}

// FiltColMap find values that matches min <= A <= max for a specific column.
//
// Return Value
//
// matches - a map[int]float64 where the key is the row number in mat,
// and the value is the value in the column specified by col.
func (A DenseMatrix) FiltColMap(min, max float64, col int) (matches map[int]float64, err error) {
	r := A.rows
	c := A.cols
	matches = make(map[int]float64)

	if col < 0 || col > c-1 {
		return matches, errors.New(fmt.Sprintf("matutil: Expected col vaule in range 0 to %d.  Received %d\n", c-1, col))
	}

	for i := 0; i < r; i++ {
		v := A.Get(i, col)
		if v >= min && v <= max {
			matches[i] = v
		}
	}
	return
}

// AppendCol appends column to an existing matrix.  If length of column
// is greater than the number of rows in the matrix, and error is returned.
// If the length of column is less than the number of rows, the column is padded
// with zeros.
//
// Returns a new matrix with the column append and leaves the source untouched.
func (A *DenseMatrix) AppendCol(column []float64) (*DenseMatrix, error) {
	rows := A.rows
	cols := A.cols
	var err error = nil
	if len(column) > rows {
		return Zeros(1, 1), errors.New(fmt.Sprintf("Cannot append a column with %d elements to an matrix with %d rows.", len(column), rows))
	}
	// Put the source array into a slice.
	// If there are R rows and C columns, the first C elements hold the data in
	// the first row, the 2nd C elements hold the data in the 2nd row, etc.
	source := make([]float64, rows*cols+len(column))
	for i := 0; i < rows; i++ {
		j := 0
		for ; j < cols; j++ {
			source[j] = A.Get(i, j)
		}
		source[j] = column[i]
	}
	return MakeDenseMatrix(source, rows, cols+1), err
}

// AppendRow appends a row to the bottom of a matrix.
func (A *DenseMatrix) AppendRow(r *DenseMatrix) (*DenseMatrix, error) {
	rows := A.rows
	cols := A.cols
	rrows := r.rows
	rcols := r.cols
	if cols != rcols {
		return Zeros(1, 1), errors.New(fmt.Sprintf("Expect %d by %d matrix.  Received %d by %d matrix.\n", rows, cols, rrows, rcols))
	}

	source := append(A.Array(), r.Array()...)
	return MakeDenseMatrix(source, rows+1, cols), nil
}

// RowExists returns true if the row vector exists in the matrix.
func (A *DenseMatrix) RowExists(r *DenseMatrix) bool {
	rows := A.rows
	cols := A.cols
	//rrows := r.rows
	rcols := r.cols
	if cols != rcols {
		return false
	}

	for i := 0; i < rows; i++ {
		rowexists := true
		for j := 0; j < cols; j++ {
			if r.Get(0, j) != A.Get(i, j) {
				rowexists = false
				break
			}
		}
		if rowexists {
			return true
		}
	}
	return false
}

// ColSlice retrieves the values in column i of a matrix as a slice
func (A DenseMatrix) ColSlice(col int) []float64 {
	rows := A.rows
	r := make([]float64, rows)
	for j := 0; j < rows; j++ {
		r[j] = A.Get(j, col)
	}
	return r
}

// SumCol calculates the sum of the indicated column and returns a float64
func (A DenseMatrix) SumCol(col int) float64 {
	//numRows _ := GetSize()
	numRows := A.rows
	sum := float64(0)

	for i := 0; i < numRows; i++ {
		sum += A.Get(i, col)
	}
	return sum
}

// SumCols takes the sum of each column in the matrix and returns a mX1 matrix of
// the sums.
func (m DenseMatrix) SumCols() *DenseMatrix {
	numRows, numCols := m.GetSize()
	sums := Zeros(1, numCols)

	for j := 0; j < numCols; j++ {
		i := 0
		s := 0.0
		for ; i < numRows; i++ {
			s += m.Get(i, j)
		}
		sums.Set(0, j, s)
	}
	return sums
}

// MeanCols calculates the mean of the columns and returns a 1Xn matrix
func (m DenseMatrix) MeanCols() *DenseMatrix {
	numRows, numCols := m.GetSize()
	sums := m.SumCols()
	means := Zeros(1, numCols)
	b := float64(0)

	for j := 0; j < numCols; j++ {
		b = sums.Get(0, j) / float64(numRows)
		means.Set(0, j, b)
	}
	return means
}

// SumRows calculates the sum of each row in a matrix and returns a nX1 matrix of
// the sums.
func (A DenseMatrix) SumRows() *DenseMatrix {
	numRows, numCols := A.GetSize()
	sums := Zeros(numRows, 1)

	for i := 0; i < numRows; i++ {
		j := 0
		s := 0.0
		for ; j < numCols; j++ {
			s += A.Get(i, j)
		}
		sums.Set(i, 0, s)
	}
	return sums
}

// SumRowsM calculates the sum of each row in a matrix and modifes the first
// column of the row so that it contains the sum.  This is for performance
// to avoid allocations made by Zeros()
func (A *DenseMatrix) SumRowsM() {
	numRows, numCols := A.GetSize()
	//sums := Zeros(numRows, 1)

	for i := 0; i < numRows; i++ {
		j := 0
		s := 0.0
		for ; j < numCols; j++ {
			s += A.Get(i, j)
		}
		A.Set(i, 0, s)
	}
}

// Arrays returns an array of slices referencing the matrix data. Changes to
// the slices effect changes to the matrix.
func (A *DenseMatrix) Arrays() [][]float64 {
	a := make([][]float64, A.rows)
	for i := 0; i < A.rows; i++ {
		a[i] = A.elements[i*A.step : i*A.step+A.cols]
	}
	return a
}

// Array returns the contents of this matrix stored into a flat array (row-major).
func (A *DenseMatrix) Array() []float64 {
	if A.step == A.rows {
		return A.elements[0 : A.rows*A.cols]
	}
	a := make([]float64, A.rows*A.cols)
	for i := 0; i < A.rows; i++ {
		for j := 0; j < A.cols; j++ {
			a[i*A.cols+j] = A.elements[i*A.step+j]
		}
	}
	return a
}

// rowSlice returns a slice at given row
func (A *DenseMatrix) rowSlice(row int) []float64 {
	return A.elements[row*A.step : row*A.step+A.cols]
}

// Get returns the element in the ith row and jth column.
func (A *DenseMatrix) Get(i int, j int) (v float64) {
	/*
		i = i % A.rows
		if i < 0 {
			i = A.rows - i
		}
		j = j % A.cols
		if j < 0 {
			j = A.cols - j
		}
	*/

	// reslicing like this does efficient range checks, perhaps
	v = A.elements[i*A.step : i*A.step+A.cols][j]
	//v = A.elements[i*A.step+j]
	return
}

// Set the element in the ith row and jth column to v.
func (A *DenseMatrix) Set(i int, j int, v float64) {
	/*
		i = i % A.rows
		if i < 0 {
			i = A.rows - i
		}
		j = j % A.cols
		if j < 0 {
			j = A.cols - j
		}
	*/
	// reslicing like this does efficient range checks, perhaps
	A.elements[i*A.step : i*A.step+A.cols][j] = v
	//A.elements[i*A.step+j] = v
}

// GetMatrix returns a submatrix starting at i,j with rows rows and cols columns. Changes to
// the returned matrix show up in the original.
func (A *DenseMatrix) GetMatrix(i, j, rows, cols int) *DenseMatrix {
	B := new(DenseMatrix)
	B.elements = A.elements[i*A.step+j : i*A.step+j+(rows-1)*A.step+cols]
	B.rows = rows
	B.cols = cols
	B.step = A.step
	return B
}

// SetMatrix copies B into A, with B's 0, 0 aligning with A's i, j
func (A *DenseMatrix) SetMatrix(i, j int, B *DenseMatrix) {
	for r := 0; r < B.rows; r++ {
		for c := 0; c < B.cols; c++ {
			A.Set(i+r, j+c, B.Get(r, c))
		}
	}
}

// GetColVector gets the matrix at j
func (A *DenseMatrix) GetColVector(j int) *DenseMatrix {
	return A.GetMatrix(0, j, A.rows, 1)
}

// GetRowVector returns the matrix at i
func (A *DenseMatrix) GetRowVector(i int) *DenseMatrix {
	return A.GetMatrix(i, 0, 1, A.cols)
}

// SetRowVector sets a row in the matrix to the values in row 0 of the
// source matrix.  If there are more columns in the source than in
// the target the target columns are filled up to number of columns.
func (A *DenseMatrix) SetRowVector(src *DenseMatrix, row int) {
	_, cols := A.GetSize()
	for i := 0; i < cols; i++ {
		A.Set(row, i, src.Get(0, i))
	}
}

/*
Get a copy of this matrix with 0s above the diagonal.
*/
func (A *DenseMatrix) L() *DenseMatrix {
	B := A.Copy()
	for i := 0; i < A.rows; i++ {
		for j := i + 1; j < A.cols; j++ {
			B.Set(i, j, 0)
		}
	}
	return B
}

// U gets a copy of this matrix with 0s below the diagonal.
func (A *DenseMatrix) U() *DenseMatrix {
	B := A.Copy()
	for i := 0; i < A.rows; i++ {
		for j := 0; j < i && j < A.cols; j++ {
			B.Set(i, j, 0)
		}
	}
	return B
}

// Copy returns a copy of the matrix A
func (A *DenseMatrix) Copy() *DenseMatrix {
	B := new(DenseMatrix)
	B.rows = A.rows
	B.cols = A.cols
	B.step = A.cols
	B.elements = make([]float64, B.rows*B.cols)
	for row := 0; row < B.rows; row++ {
		copy(B.rowSlice(row), A.rowSlice(row))
	}
	return B
}

// Augment get a new matrix [A B].
func (A *DenseMatrix) Augment(B *DenseMatrix) (C *DenseMatrix, err error) {
	if A.rows != B.rows {
		err = ErrorDimensionMismatch
		return
	}
	C = Zeros(A.rows, A.cols+B.cols)
	err = A.AugmentFill(B, C)
	return
}

// AugmentFill returns
func (A *DenseMatrix) AugmentFill(B, C *DenseMatrix) (err error) {
	if A.rows != B.rows || C.rows != A.rows || C.cols != A.cols+B.cols {
		err = ErrorDimensionMismatch
		return
	}
	C.SetMatrix(0, 0, A)
	C.SetMatrix(0, A.cols, B)
	/*
		for i := 0; i < C.Rows(); i++ {
			for j := 0; j < A.Cols(); j++ {
				C.Set(i, j, A.Get(i, j))
			}
			for j := 0; j < B.Cols(); j++ {
				C.Set(i, j+A.Cols(), B.Get(i, j))
			}
		}*/
	return
}

/*
Get a new matrix [A; B], with A above B.
*/
func (A *DenseMatrix) Stack(B *DenseMatrix) (C *DenseMatrix, err error) {
	if A.cols != B.cols {
		err = ErrorDimensionMismatch
		return
	}
	C = Zeros(A.rows+B.rows, A.cols)
	err = A.StackFill(B, C)
	return
}
func (A *DenseMatrix) StackFill(B, C *DenseMatrix) (err error) {
	if A.cols != B.cols || C.cols != A.cols || C.rows != A.rows+B.rows {
		err = ErrorDimensionMismatch
		return
	}
	C.SetMatrix(0, 0, A)
	C.SetMatrix(A.rows, 0, B)
	/*
		for j := 0; j < A.cols; j++ {
			for i := 0; i < A.Rows(); i++ {
				C.Set(i, j, A.Get(i, j))
			}
			for i := 0; i < B.cols; i++ {
				C.Set(i+A.rows, j, B.Get(i, j))
			}
		}
	*/
	return
}

/*
Create a sparse matrix copy.
*/
func (A *DenseMatrix) SparseMatrix() *SparseMatrix {
	B := ZerosSparse(A.rows, A.cols)
	for i := 0; i < A.rows; i++ {
		for j := 0; j < A.cols; j++ {
			v := A.Get(i, j)
			if v != 0 {
				B.Set(i, j, v)
			}
		}
	}
	return B
}

func (A *DenseMatrix) DenseMatrix() *DenseMatrix {
	return A.Copy()
}

func Zeros(rows, cols int) *DenseMatrix {
	A := new(DenseMatrix)
	A.elements = make([]float64, rows*cols)
	A.rows = rows
	A.cols = cols
	A.step = cols
	return A
}

func Ones(rows, cols int) *DenseMatrix {
	A := new(DenseMatrix)
	A.elements = make([]float64, rows*cols)
	A.rows = rows
	A.cols = cols
	A.step = cols

	for i := 0; i < len(A.elements); i++ {
		A.elements[i] = 1
	}

	return A
}

func Numbers(rows, cols int, num float64) *DenseMatrix {
	A := Zeros(rows, cols)

	for i := 0; i < A.Rows(); i++ {
		for j := 0; j < A.Cols(); j++ {
			A.Set(i, j, num)
		}
	}

	return A
}

/*
Create an identity matrix with span rows and span columns.
*/
func Eye(span int) *DenseMatrix {
	A := Zeros(span, span)
	for i := 0; i < span; i++ {
		A.Set(i, i, 1)
	}
	return A
}

func Normals(rows, cols int) *DenseMatrix {
	A := Zeros(rows, cols)

	for i := 0; i < A.Rows(); i++ {
		for j := 0; j < A.Cols(); j++ {
			A.Set(i, j, rand.NormFloat64())
		}
	}

	return A
}

func Diagonal(d []float64) *DenseMatrix {
	n := len(d)
	A := Zeros(n, n)
	for i := 0; i < n; i++ {
		A.Set(i, i, d[i])
	}
	return A
}

func MakeDenseCopy(A MatrixRO) *DenseMatrix {
	B := Zeros(A.Rows(), A.Cols())
	for i := 0; i < B.rows; i++ {
		for j := 0; j < B.cols; j++ {
			B.Set(i, j, A.Get(i, j))
		}
	}
	return B
}

func MakeDenseMatrix(elements []float64, rows, cols int) *DenseMatrix {
	A := new(DenseMatrix)
	A.elements = make([]float64, rows*cols)
	A.rows = rows
	A.cols = cols
	A.step = cols
	A.elements = elements
	return A
}

func MakeDenseMatrixStacked(data [][]float64) *DenseMatrix {
	rows := len(data)
	cols := len(data[0])
	elements := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			elements[i*cols+j] = data[i][j]
		}
	}
	return MakeDenseMatrix(elements, rows, cols)
}

func (A *DenseMatrix) String() string { return String(A) }

// Calculate distance matrix for a given r norm.  Use r = 2 for Euclidean.
func (A *DenseMatrix) Distance(r float64, lower bool) *DenseMatrix {
	B := Zeros(A.Cols(), A.Cols())

	ir := 1.0 / r
	for i := 0; i < A.Cols(); i++ {
		for j := i + 1; j < A.Cols(); j++ {
			colI := A.ColSlice(i)
			colJ := A.ColSlice(j)

			diffs := 0.0
			for index, _ := range colI {
				diffs += math.Pow(math.Abs(colI[index]-colJ[index]), r)
			}

			dist := math.Pow(diffs, ir)
			B.Set(i, j, dist)
			if lower {
				B.Set(j, i, dist)
			}
		}
	}
	return B
}
