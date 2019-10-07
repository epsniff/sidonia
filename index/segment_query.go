package index

import (
	"context"
	"fmt"

	"github.com/araddon/gou"

	"github.com/RoaringBitmap/roaring"
	"github.com/couchbase/vellum"
	"github.com/couchbase/vellum/regexp"
)

type Query interface {
	Type() QType
}

type QType int

const (
	TypeRegExtQuery QType = 10
)

type QueryBuilder struct {
	ctx context.Context
	seg *Segment
	ops func() (*SearchResults, error) // TODO drop SearchResults in favor of using a Roaring bitmap to gather results in.
}

func NewQueryBuilder(ctx context.Context, seg *Segment) *QueryBuilder {
	return &QueryBuilder{ctx, seg, nil}
}

func (q *QueryBuilder) And(queries ...Query) *QueryBuilder {
	currentOps := q.ops
	op := func() (*SearchResults, error) {
		// TODO and all the queries together.
		res := &SearchResults{roaring.New(), nil}
		firstRun := true
		for _, query := range queries {
			switch query.Type() {
			case TypeRegExtQuery:
				tmpq := query.(*RegExTermQuery)
				results, err := q.seg.QueryRegEx(q.ctx, tmpq)
				if err != nil {
					return nil, err
				}
				if firstRun {
					res.internalDocIds.Or(results.internalDocIds) // Add all of them on the first loop
					firstRun = false
				} else {
					res.internalDocIds.And(results.internalDocIds)
				}
				// fmt.Printf(" DEBUG >> %v  --> %v  \n", query, res.internalDocIds.ToArray())
			default:
				return nil, fmt.Errorf("unsupported query type")
			}

		}

		// TODO this isn't write, for now we're treating all children as OR statements with this batch of AND blocks
		//      We need to rethink this and consider this and decide of a better way to build an AST ?
		if currentOps != nil {
			// TODO This whole block is all wrong, but I need to revisit it later.
			children, err := currentOps()
			if err != nil {
				return nil, err
			}
			res.internalDocIds.Or(children.internalDocIds)
		}
		return res, nil
	}

	q.ops = op
	return q
}

func (q *QueryBuilder) Run() (*SearchResults, error) {
	results, err := q.ops()
	if err != nil {
		gou.Errorf("error running query: err:%v", err)
		return nil, err
	}
	// for i, internalDocID := range results.internalDocIds.ToArray() {
	// 	externalDocID, ok := q.seg.docIDInternalToExternal[internalDocID]
	// 	if !ok {
	// 		gou.Warnf("found an internal docID without an external doc ID mapping: id:%v", internalDocID)
	// 	}
	// 	array[i] = externalDocID
	// }
	array, err := GetExternalIDs(q.seg, results.internalDocIds)
	if err != nil {
		gou.Errorf("error from GetExternalIDs: err:%v", err)
		return nil, err
	}
	results.ExternalDocIDs = array

	return results, nil
}

// GetExternalIDs takes a bitmap of internal ids and converts them to an array of external ids
// extacted from the segment.
func GetExternalIDs(seg *Segment, internalDocIds *roaring.Bitmap) ([]string, error) {
	array := make([]string, internalDocIds.GetCardinality())
	postingIter := internalDocIds.Iterator()
	i := 0
	for postingIter.HasNext() {
		internalDocID := postingIter.Next()
		externalDocID, ok := seg.docIDInternalToExternal[internalDocID]
		if !ok {
			return nil, fmt.Errorf("found an internal docID without an external doc ID mapping: id:%v", internalDocID)
		}
		array[i] = externalDocID
		i++
	}
	return array, nil
}

type SearchResults struct {
	internalDocIds *roaring.Bitmap

	ExternalDocIDs []string
}

type RegExTermQuery struct {
	Fieldname string
	RegEx     string
}

func (q *RegExTermQuery) Type() QType {
	return TypeRegExtQuery
}

func (seg *Segment) QueryRegEx(ctx context.Context, query *RegExTermQuery) (*SearchResults, error) {

	field := query.Fieldname
	regEx := query.RegEx

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
		seg.termDicFstCache[fieldId] = termDictionary
	}
	//
	// Query the Term Dic
	//
	r, err := regexp.New(regEx)
	if err != nil {
		return nil, err
	}

	var res *SearchResults = &SearchResults{roaring.New(), nil}
	itr, err := termDictionary.Search(r, nil, nil)
	for ; err == nil; err = itr.Next() {
		_, termID := itr.Current()
		postingList := seg.postings[uint32(termID)]
		postings := postingList.Postings()
		res.internalDocIds.Or(postings)

		// fmt.Printf("found in term - termId:%d %s %s :: Card:%v\n", termID, field, regEx, res.internalDocIds.GetCardinality())

		// fmt.Printf("  docs list: %v  \n", seg.postings[uint32(termID)])
		// postingIter := postings.Iterator()
		// for postingIter.HasNext() {
		// 	docId := postingIter.Next()
		// 	// eDocId, ok := seg.docIDInternalToExternal[docId]
		// 	if ok {
		// 		// fmt.Printf("  found doc: %v \n", eDocId)
		// 		res.DocIds.Add // = append(res.DocIds, eDocId)
		// 	}
		// }
	}
	// fmt.Printf("  ----%v\n", seg.terminIdInt)

	return res, nil
}
