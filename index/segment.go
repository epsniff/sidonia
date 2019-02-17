package index

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/RoaringBitmap/roaring"
	"github.com/couchbase/vellum"
	"github.com/couchbase/vellum/regexp"
)

type SearchResults struct {
	DocIds []string
}

type Segment struct {
	termDicBytes []byte
	termDic      *vellum.FST

	postings map[uint32]*roaring.Bitmap // termID --> list of doc Ids // TODO replace with roaring bitmaps...

	// termID to postings list DocIds
	terminIdInt  uint32 // TODO use uint16 instead?  And limit the size of the segment to 65k terms?
	termToTermID map[string]uint32

	// docid to doc
	docIdInc                uint32
	docIDInternalToExternal map[uint32]string
	docIDExternalToInternal map[string]uint32
}

func (seg *Segment) QueryRegEx(ctx context.Context, regEx string) (*SearchResults, error) {

	if seg.termDic == nil {
		if len(seg.termDicBytes) == 0 {
			return nil, fmt.Errorf("empty term dictionary")
		}
		termDic, err := vellum.Load(seg.termDicBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to load vellem FST: %v", err)
		}
		seg.termDic = termDic
	}
	//
	// Query the Term Dic
	//
	r, err := regexp.New(regEx)
	if err != nil {
		return nil, err
	}

	var res *SearchResults = &SearchResults{}
	itr, err := seg.termDic.Search(r, nil, nil)
	for err == nil {
		_, termID := itr.Current()
		// fmt.Printf("found in term Dic: %s - termId:%d\n", term, termID)
		// fmt.Printf("  docs list: %v\n", seg.postings[uint32(termID)])
		posting := seg.postings[uint32(termID)]
		postingIter := posting.Iterator()
		for postingIter.HasNext() {
			docId := postingIter.Next()
			eDocId, ok := seg.docIDInternalToExternal[docId]
			if ok {
				// fmt.Printf("  found doc: %v \n", eDocId)
				res.DocIds = append(res.DocIds, eDocId)
			}
		}
		err = itr.Next()
	}
	// fmt.Printf("  ----%v\n", seg.terminIdInt)

	return res, nil
}

func DocsToSegment(ctx context.Context, docs []*Document) (*Segment, error) {
	index := &Segment{
		postings:                map[uint32]*roaring.Bitmap{},
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
			// fmt.Println(doc.DocID)
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
			if list, ok := index.postings[termID]; ok {
				index.postings[termID].Add(docID)
			} else {
				list = roaring.New()
				list.Add(docID)
				index.postings[termID] = list
			}
		}
	}

	sort.Sort(fields)

	//
	// Build the Term Dictionary, using an FST (vellum)
	//
	// f, err := os.Create("/tmp/term.test.dic")
	// if err != nil {
	// 	return nil, err
	// }

	buff := bytes.NewBuffer([]byte{})

	dic, err := vellum.New(buff, nil)
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
	if err := dic.Close(); err != nil {
		return nil, fmt.Errorf("vellum close failed:%v", err)
	}
	// index.termDic = dic
	index.termDicBytes = buff.Bytes()

	return index, nil
}
