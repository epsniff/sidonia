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
	// field ID
	fieldIdInt     uint32
	fieldToFieldId map[string]uint32 // external-FieldID --> internal-FieldID

	// term FieldID to Term Dic
	termDicBytes    map[uint32][]byte // internal-FieldId --> (bytes) TermDic (FST( Term -- > TermID ))
	termDicFstCache map[uint32]*vellum.FST

	// termID to postings list DocIds
	terminIdInt  uint32 // TODO use uint16 instead?  And limit the size of the segment to 65k terms?
	termToTermID map[string]uint32

	postings map[uint32]*roaring.Bitmap // termID --> list of doc Ids // TODO replace with roaring bitmaps...

	// docid to doc
	docIdInc                uint32
	docIDInternalToExternal map[uint32]string
	docIDExternalToInternal map[string]uint32
}

func (seg *Segment) QueryRegEx(ctx context.Context, field, regEx string) (*SearchResults, error) {
	var err error
	fieldId, ok := seg.fieldToFieldId[field]
	if !ok {
		return nil, fmt.Errorf("no field-id found for field: %v", field)
	}

	termDictionary, ok := seg.termDicFstCache[fieldId]
	if !ok {
		tbytes, ok := seg.termDicBytes[fieldId]
		if !ok {
			return nil, fmt.Errorf("no term dictionary found for field: %v", field)
		}
		termDictionary, err = vellum.Load(tbytes)
		if err != nil {
			return nil, fmt.Errorf("failed loading term dictionary: err:%v", err)
		}
	}
	//
	// Query the Term Dic
	//
	r, err := regexp.New(regEx)
	if err != nil {
		return nil, err
	}

	var res *SearchResults = &SearchResults{}
	itr, err := termDictionary.Search(r, nil, nil)
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
	seg := &Segment{
		fieldToFieldId: map[string]uint32{},
		// fieldIdToTermDicBuilder: map[uint32]*vellum.Builder{},
		termToTermID:            map[string]uint32{},
		termDicBytes:            map[uint32][]byte{},
		postings:                map[uint32]*roaring.Bitmap{},
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
		if did, ok := seg.docIDExternalToInternal[doc.DocID]; ok {
			docID = did
		} else {
			docID = seg.docIdInc
			seg.docIDExternalToInternal[doc.DocID] = docID
			seg.docIDInternalToExternal[docID] = doc.DocID
			// fmt.Println(doc.DocID)
			seg.docIdInc++
		}
		for field, term := range doc.Fields {
			fieldID := uint32(0)
			if fid, ok := seg.fieldToFieldId[field]; ok {
				fieldID = fid
			} else {
				fieldID = seg.fieldIdInt
				seg.fieldToFieldId[field] = fieldID
				seg.fieldIdInt++
			}

			termID := uint32(0)
			// term := fmt.Sprintf("%v::%v", key, val)
			if tid, ok := seg.termToTermID[term]; ok {
				termID = tid
			} else {
				termID = seg.terminIdInt
				seg.termToTermID[term] = termID
				seg.terminIdInt++
			}

			// TODO url encode key/vals to avoid collioitions with our split char '::' ?
			// TODO is this the best way to index strutured data ?
			if iField, ok := fields[fieldID]; ok {
				iField.Terms = append(iField.Terms, &Term{Term: term, TermID: termID})
			} else {
				iField = &IndexableField{
					InternalDocId: docID,
					FieldID:       fieldID,
					FieldName:     field,
				}
				iField.Terms = append(iField.Terms, &Term{Term: term, TermID: termID})
				fields[fieldID] = iField
			}
			// fields = append(fields, &IndexableField{InternalDocId: docID, FieldID: fieldID, Term: term, TermID: termID})
			if list, ok := seg.postings[termID]; ok {
				seg.postings[termID].Add(docID)
			} else {
				list = roaring.New()
				list.Add(docID)
				seg.postings[termID] = list
			}
		}
	}

	//
	// Build the Term Dictionary, using an FST (vellum)
	//
	// f, err := os.Create("/tmp/term.test.dic")
	// if err != nil {
	// 	return nil, err
	// }

	mkFst := func() (*vellum.Builder, *bytes.Buffer, error) {
		buff := bytes.NewBuffer([]byte{})
		var vellumOptions *vellum.BuilderOpts
		dic, err := vellum.New(buff, vellumOptions)
		if err != nil {
			return nil, nil, err
		}
		return dic, buff, nil
	}
	for _, field := range fields {
		sort.Sort(field.Terms)

		// TODO lets stop saving the map of term dics builders on index and
		// save them onto the IndexableField struct

		fst, buff, err := mkFst()
		if err != nil {
			return nil, fmt.Errorf("failed to create FST builder: %v", err)
		}
		for _, term := range field.Terms {
			err := fst.Insert([]byte(term.Term), uint64(term.TermID))
			if err != nil {
				return nil, err
			}
		}

		if err := fst.Close(); err != nil {
			return nil, fmt.Errorf("vellum close failed:%v", err)
		}
		seg.termDicBytes[field.FieldID] = buff.Bytes()
	}

	return seg, nil
}
