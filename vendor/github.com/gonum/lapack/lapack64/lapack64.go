// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lapack64 provides a set of convenient wrapper functions for LAPACK
// calls, as specified in the netlib standard (www.netlib.org).
//
// The native Go routines are used by default, and the Use function can be used
// to set an alternative implementation.
//
// If the type of matrix (General, Symmetric, etc.) is known and fixed, it is
// used in the wrapper signature. In many cases, however, the type of the matrix
// changes during the call to the routine, for example the matrix is symmetric on
// entry and is triangular on exit. In these cases the correct types should be checked
// in the documentation.
//
// The full set of Lapack functions is very large, and it is not clear that a
// full implementation is desirable, let alone feasible. Please open up an issue
// if there is a specific function you need and/or are willing to implement.
package lapack64

import (
	"github.com/gonum/blas"
	"github.com/gonum/blas/blas64"
	"github.com/gonum/lapack"
	"github.com/gonum/lapack/native"
)

var lapack64 lapack.Float64 = native.Implementation{}

// Use sets the LAPACK float64 implementation to be used by subsequent BLAS calls.
// The default implementation is native.Implementation.
func Use(l lapack.Float64) {
	lapack64 = l
}

// Potrf computes the cholesky factorization of a.
//  A = U^T * U if ul == blas.Upper
//  A = L * L^T if ul == blas.Lower
// The underlying data between the input matrix and output matrix is shared.
func Potrf(a blas64.Symmetric) (t blas64.Triangular, ok bool) {
	ok = lapack64.Dpotrf(a.Uplo, a.N, a.Data, a.Stride)
	t.Uplo = a.Uplo
	t.N = a.N
	t.Data = a.Data
	t.Stride = a.Stride
	t.Diag = blas.NonUnit
	return
}

// Gecon estimates the reciprocal of the condition number of the n×n matrix A
// given the LU decomposition of the matrix. The condition number computed may
// be based on the 1-norm or the ∞-norm.
//
// The slice a contains the result of the LU decomposition of A as computed by Dgetrf.
//
// anorm is the corresponding 1-norm or ∞-norm of the original matrix A.
//
// work is a temporary data slice of length at least 4*n and Gecon will panic otherwise.
//
// iwork is a temporary data slice of length at least n and Gecon will panic otherwise.
func Gecon(norm lapack.MatrixNorm, a blas64.General, anorm float64, work []float64, iwork []int) float64 {
	return lapack64.Dgecon(norm, a.Cols, a.Data, a.Stride, anorm, work, iwork)
}

// Gels finds a minimum-norm solution based on the matrices A and B using the
// QR or LQ factorization. Dgels returns false if the matrix
// A is singular, and true if this solution was successfully found.
//
// The minimization problem solved depends on the input parameters.
//
//  1. If m >= n and trans == blas.NoTrans, Dgels finds X such that || A*X - B||_2
//     is minimized.
//  2. If m < n and trans == blas.NoTrans, Dgels finds the minimum norm solution of
//     A * X = B.
//  3. If m >= n and trans == blas.Trans, Dgels finds the minimum norm solution of
//     A^T * X = B.
//  4. If m < n and trans == blas.Trans, Dgels finds X such that || A*X - B||_2
//     is minimized.
// Note that the least-squares solutions (cases 1 and 3) perform the minimization
// per column of B. This is not the same as finding the minimum-norm matrix.
//
// The matrix A is a general matrix of size m×n and is modified during this call.
// The input matrix B is of size max(m,n)×nrhs, and serves two purposes. On entry,
// the elements of b specify the input matrix B. B has size m×nrhs if
// trans == blas.NoTrans, and n×nrhs if trans == blas.Trans. On exit, the
// leading submatrix of b contains the solution vectors X. If trans == blas.NoTrans,
// this submatrix is of size n×nrhs, and of size m×nrhs otherwise.
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= max(m,n) + max(m,n,nrhs), and this function will panic
// otherwise. A longer work will enable blocked algorithms to be called.
// In the special case that lwork == -1, work[0] will be set to the optimal working
// length.
func Gels(trans blas.Transpose, a blas64.General, b blas64.General, work []float64, lwork int) bool {
	return lapack64.Dgels(trans, a.Rows, a.Cols, b.Cols, a.Data, a.Stride, b.Data, b.Stride, work, lwork)
}

// Geqrf computes the QR factorization of the m×n matrix A using a blocked
// algorithm. A is modified to contain the information to construct Q and R.
// The upper triangle of a contains the matrix R. The lower triangular elements
// (not including the diagonal) contain the elementary reflectors. Tau is modified
// to contain the reflector scales. Tau must have length at least min(m,n), and
// this function will panic otherwise.
//
// The ith elementary reflector can be explicitly constructed by first extracting
// the
//  v[j] = 0           j < i
//  v[j] = i           j == i
//  v[j] = a[i*lda+j]  j > i
// and computing h_i = I - tau[i] * v * v^T.
//
// The orthonormal matrix Q can be constucted from a product of these elementary
// reflectors, Q = H_1*H_2 ... H_k, where k = min(m,n).
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= m and this function will panic otherwise.
// Dgeqrf is a blocked QR factorization, but the block size is limited
// by the temporary space available. If lwork == -1, instead of performing Geqrf,
// the optimal work length will be stored into work[0].
func Geqrf(a blas64.General, tau, work []float64, lwork int) {
	lapack64.Dgeqrf(a.Rows, a.Cols, a.Data, a.Stride, tau, work, lwork)
}

// Gelqf computes the QR factorization of the m×n matrix A using a blocked
// algorithm. A is modified to contain the information to construct L and Q.
// The lower triangle of a contains the matrix L. The lower triangular elements
// (not including the diagonal) contain the elementary reflectors. Tau is modified
// to contain the reflector scales. Tau must have length at least min(m,n), and
// this function will panic otherwise.
//
// See Geqrf for a description of the elementary reflectors and orthonormal
// matrix Q. Q is constructed as a product of these elementary reflectors,
// Q = H_k ... H_2*H_1.
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= m and this function will panic otherwise.
// Dgeqrf is a blocked LQ factorization, but the block size is limited
// by the temporary space available. If lwork == -1, instead of performing Gelqf,
// the optimal work length will be stored into work[0].
func Gelqf(a blas64.General, tau, work []float64, lwork int) {
	lapack64.Dgelqf(a.Rows, a.Cols, a.Data, a.Stride, tau, work, lwork)
}

// Getrf computes the LU decomposition of the m×n matrix A.
// The LU decomposition is a factorization of A into
//  A = P * L * U
// where P is a permutation matrix, L is a unit lower triangular matrix, and
// U is a (usually) non-unit upper triangular matrix. On exit, L and U are stored
// in place into a.
//
// ipiv is a permutation vector. It indicates that row i of the matrix was
// changed with ipiv[i]. ipiv must have length at least min(m,n), and will panic
// otherwise. ipiv is zero-indexed.
//
// Dgetrf is the blocked version of the algorithm.
//
// Dgetrf returns whether the matrix A is singular. The LU decomposition will
// be computed regardless of the singularity of A, but division by zero
// will occur if the false is returned and the result is used to solve a
// system of equations.
func Getrf(a blas64.General, ipiv []int) bool {
	return lapack64.Dgetrf(a.Rows, a.Cols, a.Data, a.Stride, ipiv)
}

// Getri computes the inverse of the matrix A using the LU factorization computed
// by Getrf. On entry, a contains the PLU decomposition of A as computed by
// Getrf and on exit contains the reciprocal of the original matrix.
//
// Getri will not perform the inversion if the matrix is singular, and returns
// a boolean indicating whether the inversion was successful.
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= n and this function will panic otherwise.
// Dgetri is a blocked inversion, but the block size is limited
// by the temporary space available. If lwork == -1, instead of performing Getri,
// the optimal work length will be stored into work[0].
func Getri(a blas64.General, ipiv []int, work []float64, lwork int) (ok bool) {
	return lapack64.Dgetri(a.Cols, a.Data, a.Stride, ipiv, work, lwork)
}

// Dgetrs solves a system of equations using an LU factorization.
// The system of equations solved is
//  A * X = B if trans == blas.Trans
//  A^T * X = B if trans == blas.NoTrans
// A is a general n×n matrix with stride lda. B is a general matrix of size n×nrhs.
//
// On entry b contains the elements of the matrix B. On exit, b contains the
// elements of X, the solution to the system of equations.
//
// a and ipiv contain the LU factorization of A and the permutation indices as
// computed by Getrf. ipiv is zero-indexed.
func Getrs(trans blas.Transpose, a blas64.General, b blas64.General, ipiv []int) {
	lapack64.Dgetrs(trans, a.Cols, b.Cols, a.Data, a.Stride, ipiv, b.Data, b.Stride)
}

// Lange computes the matrix norm of the general m×n matrix A. The input norm
// specifies the norm computed.
//  lapack.MaxAbs: the maximum absolute value of an element.
//  lapack.MaxColumnSum: the maximum column sum of the absolute values of the entries.
//  lapack.MaxRowSum: the maximum row sum of the absolute values of the entries.
//  lapack.Frobenius: the square root of the sum of the squares of the entries.
// If norm == lapack.MaxColumnSum, work must be of length n, and this function will panic otherwise.
// There are no restrictions on work for the other matrix norms.
func Lange(norm lapack.MatrixNorm, a blas64.General, work []float64) float64 {
	return lapack64.Dlange(norm, a.Rows, a.Cols, a.Data, a.Stride, work)
}

// Lansy computes the specified norm of an n×n symmetric matrix. If
// norm == lapack.MaxColumnSum or norm == lapackMaxRowSum work must have length
// at least n and this function will panic otherwise.
// There are no restrictions on work for the other matrix norms.
func Lansy(norm lapack.MatrixNorm, a blas64.Symmetric, work []float64) float64 {
	return lapack64.Dlansy(norm, a.Uplo, a.N, a.Data, a.Stride, work)
}

// Lantr computes the specified norm of an m×n trapezoidal matrix A. If
// norm == lapack.MaxColumnSum work must have length at least n and this function
// will panic otherwise. There are no restrictions on work for the other matrix norms.
func Lantr(norm lapack.MatrixNorm, a blas64.Triangular, work []float64) float64 {
	return lapack64.Dlantr(norm, a.Uplo, a.Diag, a.N, a.N, a.Data, a.Stride, work)
}

// Ormlq multiplies the matrix C by the othogonal matrix Q defined by
// A and tau. A and tau are as returned from Gelqf.
//  C = Q * C    if side == blas.Left and trans == blas.NoTrans
//  C = Q^T * C  if side == blas.Left and trans == blas.Trans
//  C = C * Q    if side == blas.Right and trans == blas.NoTrans
//  C = C * Q^T  if side == blas.Right and trans == blas.Trans
// If side == blas.Left, A is a matrix of side k×m, and if side == blas.Right
// A is of size k×n. This uses a blocked algorithm.
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= m if side == blas.Left and lwork >= n if side == blas.Right,
// and this function will panic otherwise.
// Ormlq uses a block algorithm, but the block size is limited
// by the temporary space available. If lwork == -1, instead of performing Ormlq,
// the optimal work length will be stored into work[0].
//
// Tau contains the householder scales and must have length at least k, and
// this function will panic otherwise.
func Ormlq(side blas.Side, trans blas.Transpose, a blas64.General, tau []float64, c blas64.General, work []float64, lwork int) {
	lapack64.Dormlq(side, trans, c.Rows, c.Cols, a.Rows, a.Data, a.Stride, tau, c.Data, c.Stride, work, lwork)
}

// Ormqr multiplies the matrix C by the othogonal matrix Q defined by
// A and tau. A and tau are as returned from Geqrf.
//  C = Q * C    if side == blas.Left and trans == blas.NoTrans
//  C = Q^T * C  if side == blas.Left and trans == blas.Trans
//  C = C * Q    if side == blas.Right and trans == blas.NoTrans
//  C = C * Q^T  if side == blas.Right and trans == blas.Trans
// If side == blas.Left, A is a matrix of side k×m, and if side == blas.Right
// A is of size k×n. This uses a blocked algorithm.
//
// tau contains the householder scales and must have length at least k, and
// this function will panic otherwise.
//
// Work is temporary storage, and lwork specifies the usable memory length.
// At minimum, lwork >= m if side == blas.Left and lwork >= n if side == blas.Right,
// and this function will panic otherwise.
// Ormqr uses a block algorithm, but the block size is limited
// by the temporary space available. If lwork == -1, instead of performing Ormqr,
// the optimal work length will be stored into work[0].
func Ormqr(side blas.Side, trans blas.Transpose, a blas64.General, tau []float64, c blas64.General, work []float64, lwork int) {
	lapack64.Dormqr(side, trans, c.Rows, c.Cols, a.Cols, a.Data, a.Stride, tau, c.Data, c.Stride, work, lwork)
}

// Pocon estimates the reciprocal of the condition number of a positive-definite
// matrix A given the Cholesky decmposition of A. The condition number computed
// is based on the 1-norm and the ∞-norm.
//
// anorm is the 1-norm and the ∞-norm of the original matrix A.
//
// work is a temporary data slice of length at least 3*n and Pocon will panic otherwise.
//
// iwork is a temporary data slice of length at least n and Pocon will panic otherwise.
func Pocon(a blas64.Symmetric, anorm float64, work []float64, iwork []int) float64 {
	return lapack64.Dpocon(a.Uplo, a.N, a.Data, a.Stride, anorm, work, iwork)
}

// Trcon estimates the reciprocal of the condition number of a triangular matrix A.
// The condition number computed may be based on the 1-norm or the ∞-norm.
//
// work is a temporary data slice of length at least 3*n and Trcon will panic otherwise.
//
// iwork is a temporary data slice of length at least n and Trcon will panic otherwise.
func Trcon(norm lapack.MatrixNorm, a blas64.Triangular, work []float64, iwork []int) float64 {
	return lapack64.Dtrcon(norm, a.Uplo, a.Diag, a.N, a.Data, a.Stride, work, iwork)
}

// Trtri computes the inverse of a triangular matrix, storing the result in place
// into a.
//
// Trtri will not perform the inversion if the matrix is singular, and returns
// a boolean indicating whether the inversion was successful.
func Trtri(a blas64.Triangular) (ok bool) {
	return lapack64.Dtrtri(a.Uplo, a.Diag, a.N, a.Data, a.Stride)
}

// Trtrs solves a triangular system of the form A * X = B or A^T * X = B. Trtrs
// returns whether the solve completed successfully. If A is singular, no solve is performed.
func Trtrs(trans blas.Transpose, a blas64.Triangular, b blas64.General) (ok bool) {
	return lapack64.Dtrtrs(a.Uplo, trans, a.Diag, a.N, b.Cols, a.Data, a.Stride, b.Data, b.Stride)
}
