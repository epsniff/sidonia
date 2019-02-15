package index

import (
	"fmt"
	"hash/fnv"
	"testing"

	"github.com/couchbase/vellum"
	"github.com/couchbase/vellum/regexp"
)

func TestIndex(t *testing.T) {
	docs := []*Document{}
	for i := 0; i < 64000; i++ {
		doc := &Document{DocID: fmt.Sprintf("%000d", i), Fields: map[string]string{}}
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
	index, err := DocsToIndex(docs)
	if err != nil {
		t.Fatalf("err:%v", err)
	}

	//
	// Query the Term Dic
	//
	dicFst, err := vellum.Open("/tmp/term.test.dic")
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	r, err := regexp.New("name::j.*")
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	itr, err := dicFst.Search(r, nil, nil)
	for err == nil {
		term, termID := itr.Current()
		fmt.Printf("found in term Dic: %s - %d\n", term, termID)
		fmt.Printf("  docs list: %v\n", index.postings[uint32(termID)])
		err = itr.Next()
	}
	fmt.Printf("  ----%v\n", index.terminIdInt)

}

func hash(s string) string {
	h := fnv.New64() // FNV hash name to int
	h.Write([]byte(s))
	key := h.Sum64()
	return fmt.Sprintf("%v", key)
}
