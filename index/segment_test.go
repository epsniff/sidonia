package index

import (
	"context"
	"fmt"
	"hash/fnv"
	"testing"
)

func TestIndex(t *testing.T) {
	docs := []*Document{}
	for i := 0; i < 64000; i++ {
		doc := &Document{DocID: fmt.Sprintf("doc_number:%000d", i), Fields: map[string]string{}}
		doc.Fields["doc_id"] = fmt.Sprintf("%000d", i)
		doc.Fields["userid"] = hash(fmt.Sprintf("%000d", i))
		switch {
		case i%100 == 0:
			doc.Fields["name"] = "eric"
		case i%100 == 1:
			doc.Fields["name"] = "kevin"
		case i%100 == 2:
			doc.Fields["name"] = "angela"
		case i%100 == 3:
			doc.Fields["name"] = "jon"
		case i%100 == 4:
			doc.Fields["name"] = "john"
		case i%100 == 5:
			doc.Fields["name"] = "james"
		}
		docs = append(docs, doc)
	}
	segment, err := DocsToSegment(context.TODO(), docs)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	res, err := segment.QueryRegEx(context.TODO(), "name::kev.*")
	fmt.Printf("found Docs: %v \n", res.DocIds)
}

func hash(s string) string {
	h := fnv.New64() // FNV hash name to int
	h.Write([]byte(s))
	key := h.Sum64()
	return fmt.Sprintf("%v", key)
}
