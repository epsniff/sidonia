package index

import "github.com/RoaringBitmap/roaring"

type indexPostings struct {
	termToPostings map[int][]int
}

type TermPostingList struct {
	TermFrequency uint32
	postings      *roaring.Bitmap
}

func (p TermPostingList) Postings() *roaring.Bitmap {
	return p.postings
}
