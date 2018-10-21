// Copyright 2009 The GoMatrix Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import "math/rand"
import "fmt"

/*
A sparse matrix based on go's map datastructure.
*/
type SparseMatrix struct {
	matrix
	elements map[int]float64
	// offset to start of matrix s.t. idx = i*cols + j + offset
	// offset = starting row * step + starting col
	offset int
	// analogous to dense step
	step int
}

func copyElements(elements map[int]float64) map[int]float64 {
	newelements := make(map[int]float64, len(elements))
	for k, v := range elements {
		newelements[k] = v
	}
	return newelements
}

func (A *SparseMatrix) Get(i, j int) float64 {
	i = i % A.rows
	if i < 0 {
		i = A.rows - i
	}
	j = j % A.cols
	if j < 0 {
		j = A.cols - j
	}
	x, _ := A.elements[i*A.step+j+A.offset]
	return x
}

/*
Looks up an element given its element index.
*/
func (A *SparseMatrix) GetIndex(index int) float64 {
	x, ok := A.elements[index]
	if !ok {
		return 0
	}
	return x
}

/*
Turn an element index into a row number.
*/
func (A *SparseMatrix) GetRowIndex(index int) (i int) {
	i = (index - A.offset) / A.cols
	return
}

/*
Turn an element index into a column number.
*/
func (A *SparseMatrix) GetColIndex(index int) (j int) {
	j = (index - A.offset) % A.cols
	return
}

/*
Turn an element index into a row and column number.
*/
func (A *SparseMatrix) GetRowColIndex(index int) (i int, j int) {
	//fmt.Printf("index: %v offset: %v step: %v i: %v\n",index,A.offset,A.step, (index-A.offset)/A.step)
	i = (index - A.offset) / A.step
	j = (index - A.offset) % A.step
	return
}

func (A *SparseMatrix) Set(i int, j int, v float64) {
	i = i % A.rows
	if i < 0 {
		i = A.rows - i
	}
	j = j % A.cols
	if j < 0 {
		j = A.cols - j
	}
	// v == 0 results in removal of key from underlying map
	if v == 0 {
		delete(A.elements, i*A.step+j+A.offset)
	} else {
		A.elements[i*A.step+j+A.offset] = v
	}
}

/*
Sets an element given its index.
*/
func (A *SparseMatrix) SetIndex(index int, v float64) {
	// v == 0 results in removal of key from underlying map
	if v == 0 {
		delete(A.elements, index)
	} else {
		A.elements[index] = v
	}
}

/*
A channel that will carry the indices of non-zero elements.
*/
func (A *SparseMatrix) Indices() (out chan int) {
	//maybe thread the populating?
	out = make(chan int)
	go func(o chan int) {
		for index := range A.elements {
			o <- index
		}
		close(o)
	}(out)
	return
}

/*
Get a matrix representing a subportion of A. Changes to the new matrix will be
reflected in A.
*/
func (A *SparseMatrix) GetMatrix(i, j, rows, cols int) (subMatrix *SparseMatrix) {
	if i < 0 || j < 0 || i+rows > A.rows || j+cols > A.cols {
		i = maxInt(0, i)
		j = maxInt(0, j)
		rows = minInt(A.rows-i, rows)
		cols = minInt(A.cols-j, cols)
	}

	subMatrix = new(SparseMatrix)
	subMatrix.rows = rows
	subMatrix.cols = cols
	subMatrix.offset = (i+A.offset/A.step)*A.step + (j + A.offset%A.step)
	//fmt.Printf("Setting offset in submatrix i:%v j:%v to %v\n",i,j,subMatrix.offset)
	subMatrix.step = A.step
	subMatrix.elements = copyElements(A.elements)
	for index, _ := range subMatrix.elements {
		ii, jj := subMatrix.GetRowColIndex(index)
		//if (ii<i || ii>i+rows || jj<j || jj>j+cols){
		//if ((ii<0 || ii>rows) && (jj<0 || jj>cols)){
		if ii < 0 || ii >= rows || jj < 0 || jj >= cols {
			delete(subMatrix.elements, index)
		}
	}
	//return subMatrix
	return
}

/*
Gets a reference to a column vector.
*/
func (A *SparseMatrix) GetColVector(j int) *SparseMatrix {
	return A.GetMatrix(0, j, A.rows, 1)
}

/*
Gets a reference to a row vector.
*/
func (A *SparseMatrix) GetRowVector(i int) *SparseMatrix {
	return A.GetMatrix(i, 0, 1, A.cols)
}

/*
Creates a new matrix [A B].
*/
func (A *SparseMatrix) Augment(B *SparseMatrix) (*SparseMatrix, error) {
	if A.rows != B.rows {
		return nil, ErrorDimensionMismatch
	}
	C := ZerosSparse(A.rows, A.cols+B.cols)

	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		C.Set(i, j, value)
	}

	for index, value := range B.elements {
		i, j := B.GetRowColIndex(index)
		C.Set(i, j+A.cols, value)
	}

	return C, nil
}

/*
Creates a new matrix [A;B], where A is above B.
*/
func (A *SparseMatrix) Stack(B *SparseMatrix) (*SparseMatrix, error) {
	if A.cols != B.cols {
		return nil, ErrorDimensionMismatch
	}
	C := ZerosSparse(A.rows+B.rows, A.cols)

	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		C.Set(i, j, value)
	}

	for index, value := range B.elements {
		i, j := B.GetRowColIndex(index)
		C.Set(i+A.rows, j, value)
	}

	return C, nil
}

/*
Returns a copy with all zeros above the diagonal.
*/
func (A *SparseMatrix) L() *SparseMatrix {
	B := ZerosSparse(A.rows, A.cols)
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i >= j {
			B.Set(i, j, value)
		}
	}
	return B
}

/*
Returns a copy with all zeros below the diagonal.
*/
func (A *SparseMatrix) U() *SparseMatrix {
	B := ZerosSparse(A.rows, A.cols)
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i <= j {
			B.Set(i, j, value)
		}
	}
	return B
}

func (A *SparseMatrix) Copy() *SparseMatrix {
	//B := ZerosSparse(A.rows, A.cols)
	B := ZerosSparseTrueCopy(A.rows, A.cols, A.offset, A.step)
	for index, value := range A.elements {
		B.elements[index] = value
	}
	return B
}
func ZerosSparseTrueCopy(rows int, cols int, offset int, step int) *SparseMatrix {
	A := new(SparseMatrix)
	A.rows = rows
	A.cols = cols
	A.offset = offset
	A.step = step
	A.elements = map[int]float64{}
	return A
}

func ZerosSparse(rows int, cols int) *SparseMatrix {
	A := new(SparseMatrix)
	A.rows = rows
	A.cols = cols
	A.offset = 0
	A.step = cols
	A.elements = map[int]float64{}
	return A
}

/*
Creates a matrix and puts a standard normal in n random elements, with replacement.
*/
func NormalsSparse(rows int, cols int, n int) *SparseMatrix {
	A := ZerosSparse(rows, cols)
	for k := 0; k < n; k++ {
		i := rand.Intn(rows)
		j := rand.Intn(cols)
		A.Set(i, j, rand.NormFloat64())
	}
	return A
}

/*
Create a sparse matrix using the provided map as its backing.
*/
func MakeSparseMatrix(elements map[int]float64, rows int, cols int) *SparseMatrix {
	A := ZerosSparse(rows, cols)
	A.elements = elements
	return A
}

/*
Convert this sparse matrix into a dense matrix.
*/
func (A *SparseMatrix) DenseMatrix() *DenseMatrix {
	B := Zeros(A.rows, A.cols)
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		B.Set(i, j, value)
	}
	return B
}

func (A *SparseMatrix) SparseMatrix() *SparseMatrix {
	return A.Copy()
}

func (A *SparseMatrix) String() string {
	return fmt.Sprintf("elements: %v step: %v offset: [%v] REST: %v", A.elements, A.step, A.offset, String(A))
}
