package matrix

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

type serializableDenseMatrix struct {
	Rows     int
	Cols     int
	Elements []float64
	Step     int
}

func (A *DenseMatrix) MarshalJSON() ([]byte, error) {
	return json.Marshal(&serializableDenseMatrix{
		Rows:     A.Rows(),
		Cols:     A.Cols(),
		Elements: A.elements,
		Step:     A.step,
	})
}

func (A *DenseMatrix) UnmarshalJSON(b []byte) error {
	tmp := &serializableDenseMatrix{}
	err := json.Unmarshal(b, tmp)
	if err != nil {
		return err
	}

	A.elements = make([]float64, tmp.Rows*tmp.Cols)

	A.rows = tmp.Rows
	A.cols = tmp.Cols
	A.elements = tmp.Elements
	A.step = tmp.Step

	return nil
}

func (A *DenseMatrix) GobEncode() ([]byte, error) {
	var b bytes.Buffer

	tmp := &serializableDenseMatrix{
		Rows:     A.Rows(),
		Cols:     A.Cols(),
		Elements: A.elements,
		Step:     A.step,
	}

	err := gob.NewEncoder(&b).Encode(tmp)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (A *DenseMatrix) GobDecode(b []byte) error {
	buf := bytes.NewBuffer(b)
	tmp := &serializableDenseMatrix{}
	err := gob.NewDecoder(buf).Decode(tmp)
	if err != nil {
		return err
	}

	A.elements = make([]float64, tmp.Rows*tmp.Cols)

	A.rows = tmp.Rows
	A.cols = tmp.Cols
	A.elements = tmp.Elements
	A.step = tmp.Step

	return nil
}
