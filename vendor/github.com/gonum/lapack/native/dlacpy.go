// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package native

import "github.com/gonum/blas"

// Dlacpy copies the elements of A specified by uplo into B. Uplo can specify
// a triangular portion with blas.Upper or blas.Lower, or can specify all of the
// elemest with blas.All.
func (impl Implementation) Dlacpy(uplo blas.Uplo, m, n int, a []float64, lda int, b []float64, ldb int) {
	checkMatrix(m, n, a, lda)
	checkMatrix(m, n, b, ldb)
	switch uplo {
	default:
		panic(badUplo)
	case blas.Upper:
		for i := 0; i < m; i++ {
			for j := i; j < n; j++ {
				b[i*ldb+j] = a[i*lda+j]
			}
		}
		return
	case blas.Lower:
		for i := 0; i < m; i++ {
			for j := 0; j < min(i, n); j++ {
				b[i*ldb+j] = a[i*lda+j]
			}
		}
		return
	case blas.All:
		for i := 0; i < m; i++ {
			for j := 0; j < n; j++ {
				b[i*ldb+j] = a[i*lda+j]
			}
		}
		return
	}
}
