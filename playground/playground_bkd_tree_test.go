package playground

import (
	"fmt"
	"testing" // "github.com/deepfabric/bkdtree"

	"github.com/epsniff/sidonia/index/bkdtree"
)

func TestBKDTree(t *testing.T) {
	t0mCap := 1000
	treesCap := 5
	bkdCap := t0mCap<<uint(treesCap) - 1
	leafCap := 50
	intraCap := 4
	numDims := 1
	bytesPerDim := 4
	dir := "/tmp"
	prefix := "bkd"
	bkd, err := bkdtree.NewBkdTree(t0mCap, leafCap, intraCap, numDims, bytesPerDim, dir, prefix)
	if err != nil {
		return
	}
	//fmt.Printf("created BkdTree %v\n", bkd)
	err = bkd.Insert(bkdtree.Point{[]uint64{55}, 55})
	if err != nil {
		t.Fatalf(" -- %v", err)
	}
	err = bkd.Insert(bkdtree.Point{[]uint64{555}, 55})
	if err != nil {
		t.Fatalf(" -- %v", err)
	}

	size := bkdCap

	lowPoint := bkdtree.Point{[]uint64{55}, 0}
	highPoint := bkdtree.Point{[]uint64{55}, 0}
	visitor := &bkdtree.IntersectCollector{lowPoint, highPoint, make([]bkdtree.Point, 0, size)}
	bkd.Intersect(visitor)

	for _, p := range visitor.Points {
		fmt.Printf(" found:%v\n", p)
	}

}
