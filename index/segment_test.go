package index

import (
	"context"
	"fmt"
	"hash/fnv"
	"testing"
	"time"

	"github.com/araddon/qlbridge/value"
	"github.com/bmizerany/assert"
)

func TestIndex(t *testing.T) {
	docs := []Document{}
	now := time.Now()
	for i := 0; i < 500; i++ {
		fieldvals := map[string]value.Value{}
		fieldvals["doc_id"] = NewStringVal(fmt.Sprintf("%000d", i))
		fieldvals["userid"] = NewStringVal(hash(fmt.Sprintf("%000d", i)))

		switch {
		case i%100 == 0:
			fieldvals["first_name"] = NewStringVal("eric")
		case i%100 == 1:
			fieldvals["first_name"] = NewStringVal("kevin")
		case i%100 == 2:
			fieldvals["first_name"] = NewStringVal("angela")
		case i%100 == 3:
			fieldvals["first_name"] = NewStringVal("jon")
		case i%100 == 4:
			fieldvals["first_name"] = NewStringVal("john")
		case i%100 == 5:
			fieldvals["first_name"] = NewStringVal("james")
		}

		if i == 101 {
			fieldvals["first_name"] = NewStringVal("kevin")
			fieldvals["last_name"] = NewStringVal("manning")
		}
		if i == 102 {
			fieldvals["last_name"] = NewStringVal("smith")
		}

		docs = append(docs, NewDocument(fmt.Sprintf("doc_number:%000d", i), fieldvals, now))
	}
	segment := NewSegment()
	err := segment.IndexDocuments(context.TODO(), docs)
	if err != nil {
		t.Fatalf("err:%v", err)
	}

	// res, err := segment.QueryRegEx(context.TODO(), &RegExTermQuery{"last_name", "man.*"})
	// if err != nil {
	// 	t.Fatalf("err:%v", err)
	// }
	// externalDocIDs, err := GetExternalIDs(segment, res.internalDocIds)
	// if err != nil {
	// 	t.Fatalf("err:%v", err)
	// }
	// fmt.Printf("found Docs: %v \n", externalDocIDs)

	//	{ // test case - invalid key
	//		_, err := NewQueryBuilder(context.TODO(), segment).
	//			And(&RegExTermQuery{"name", "kev.*"}).
	//			Run()
	//		if err == nil || !strings.Contains(err.Error(), "no field-id found for field: name") {
	//			t.Fatalf("err:%v", err)
	//		}
	//	}

	{ // test case - And regex query for first name kev.* and last_name manning
		res, err := NewQueryBuilder(context.TODO(), segment).
			And(&RegExTermQuery{"first_name", "kev.*"}, &RegExTermQuery{"last_name", "manning"}).
			Run()
		if err != nil {
			t.Fatalf("err:%v", err)
		}
		fmt.Printf("found Docs: %v \n", res.ExternalDocIDs)
		assert.Equalf(t, 1, len(res.ExternalDocIDs), "expected only one doc to match `kevin manning`")
	}

}

func hash(s string) string {
	h := fnv.New64() // FNV hash name to int
	h.Write([]byte(s))
	key := h.Sum64()
	return fmt.Sprintf("%v", key)
}
