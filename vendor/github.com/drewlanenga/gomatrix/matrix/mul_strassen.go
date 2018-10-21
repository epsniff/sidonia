// Copyright 2012 Harry de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// modified by Ralph Yozzo to support sparse matrices.

package matrix

//import "fmt"

// MulStrassen returns A * B.
//
// Original paper: Gaussian Elimination is not Optimal.
//                 Volker Strassen, 1969.
//
// This implementation is not optimized, it serves as a reference for testing.
func MulStrassen(A, B *SparseMatrix) *SparseMatrix {
	Z := ZerosSparse(A.rows, B.cols)
	C := Z.MulStrassen(A, B)
	return C
}

/*  */
// MulStrassen calculates C = A * B and returns C.
func (C *SparseMatrix) MulStrassen(A, B *SparseMatrix) *SparseMatrix {
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

	M1 := MulStrassen(A11.PlusSparseQuiet(A22), B11.PlusSparseQuiet(B22))
	M2 := MulStrassen(A21.PlusSparseQuiet(A22), B11)
	M3 := MulStrassen(A11, B12.MinusSparseQuiet(B22))
	M4 := MulStrassen(A22, B21.MinusSparseQuiet(B11))
	M5 := MulStrassen(A11.PlusSparseQuiet(A12), B22)
	M6 := MulStrassen(A21.MinusSparseQuiet(A11), B11.PlusSparseQuiet(B12))
	M7 := MulStrassen(A12.MinusSparseQuiet(A22), B21.PlusSparseQuiet(B22))

	C11.AddSparse(M7)
	C11.AddSparse(M1)
	C11.AddSparse(M4)
	C11.SubtractSparse(M5)

	C12.AddSparse(M5)
	C12.AddSparse(M3)

	C21.AddSparse(M4)
	C21.AddSparse(M2)

	C22.AddSparse(M6)
	C22.AddSparse(M1)
	C22.SubtractSparse(M2)
	C22.AddSparse(M3)
	return C
}

/* */
