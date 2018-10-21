// Copyright 2012 Harry de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// modified by Ralph Yozzo to support sparse matrices.

package matrix

//import "fmt"

// MulSimple returns A * B.
//
//
// This implementation is not optimized, it serves as a reference for testing.
func MulSimple(A, B *SparseMatrix) *SparseMatrix {
	Z := ZerosSparse(A.rows, B.cols)
	C := Z.MulSimple(A, B)
	return C
}

/*  */
// MulSimple calculates C = A * B and returns C.
func (C *SparseMatrix) MulSimple(A, B *SparseMatrix) *SparseMatrix {
	//A = A.Copy()
	//B = B.Copy()

	if A.cols < 2 {
		RETB := ZerosSparse(1, 1)
		RETB.Set(0, 0, A.Get(0, 0)*B.Get(0, 0))
		return RETB
	}
	/*
		if A.cols < 80 || A.rows != A.cols || A.rows % 2 != 0 {
			return C.MulBLAS(A, B)
		}
	*/

	m := A.rows / 2

	A11 := A.Copy().GetMatrix(0, 0, m, m)
	A12 := A.Copy().GetMatrix(0, m, m, m)
	A21 := A.Copy().GetMatrix(m, 0, m, m)
	A22 := A.Copy().GetMatrix(m, m, m, m)
	B11 := B.Copy().GetMatrix(0, 0, m, m)
	B12 := B.Copy().GetMatrix(0, m, m, m)
	B21 := B.Copy().GetMatrix(m, 0, m, m)
	B22 := B.Copy().GetMatrix(m, m, m, m)

	C11 := C.GetMatrix(0, 0, m, m)
	C12 := C.GetMatrix(0, m, m, m)
	C21 := C.GetMatrix(m, 0, m, m)
	C22 := C.GetMatrix(m, m, m, m)

	C11.AddSparse(MulSimple(A11, B11))
	C11.AddSparse(MulSimple(A12, B21))
	C12.AddSparse(MulSimple(A11, B12))
	C12.AddSparse(MulSimple(A12, B22))

	C21.AddSparse(MulSimple(A21, B11))
	C21.AddSparse(MulSimple(A22, B21))
	C22.AddSparse(MulSimple(A21, B12))
	C22.AddSparse(MulSimple(A22, B22))

	return C
}

/* */
