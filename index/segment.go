package index

import (
	"fmt"
	"os"
	"sort"

	"github.com/couchbase/vellum"
)

type Index struct {
	postings map[uint32][]uint32 // termID --> list of doc Ids // TODO replace with roaring bitmaps...

	// termID to postings list DocIds
	terminIdInt  uint32 // TODO use uint16 instead?  And limit the size of the segment to 65k terms?
	termToTermID map[string]uint32

	// docid to doc
	docIdInc                uint32
	docIDInternalToExternal map[uint32]string
	docIDExternalToInternal map[string]uint32
}

func DocsToIndex(docs []*Document) (*Index, error) {
	index := &Index{
		postings:                map[uint32][]uint32{},
		termToTermID:            map[string]uint32{},
		docIDInternalToExternal: map[uint32]string{},
		docIDExternalToInternal: map[string]uint32{},
	}

	// TODO for performance pass in a count of docs * fields so we can presize the array?
	// TODO is it faster to count the terms first, so the array can be an exact size
	fields := make(IndexableFields, 0)

	for _, doc := range docs {
		// TODO For performance we'll assume that each docId is unique?  That way we can just increment
		//  the docIdInc counter on each doc without checking if the doc already exists.
		docID := uint32(0)
		if did, ok := index.docIDExternalToInternal[doc.DocID]; ok {
			docID = did
		} else {
			docID = index.docIdInc
			index.docIDExternalToInternal[doc.DocID] = docID
			index.docIDInternalToExternal[docID] = doc.DocID
			index.docIdInc++
		}
		for key, val := range doc.Fields {
			termID := uint32(0)
			term := fmt.Sprintf("%v::%v", key, val)
			if tid, ok := index.termToTermID[term]; ok {
				termID = tid
			} else {
				termID = index.terminIdInt
				index.termToTermID[term] = termID
				index.terminIdInt++
			}
			// TODO url encode key/vals to avoid collioitions with our split char '::' ?
			// TODO is this the best way to index strutured data ?
			fields = append(fields, &IndexableField{InternalDocId: docID, Term: term, TermID: termID})
			index.postings[termID] = append(index.postings[termID], docID)
		}
	}

	sort.Sort(fields)

	//
	// Build the Term Dictionary, using an FST (vellum)
	//
	f, err := os.Create("/tmp/term.test.dic")
	if err != nil {
		return nil, err
	}

	dic, err := vellum.New(f, nil)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		// fmt.Println(field)
		err = dic.Insert([]byte(field.Term), uint64(field.TermID))
		if err != nil {
			return nil, err
		}
	}
	dic.Close()

	return index, nil
}
