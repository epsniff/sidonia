package matrix

/*
type SparseMatrix struct {
	matrix
	elements map[int]float64
	// offset to start of matrix s.t. idx = i*cols + j + offset
	// offset = starting row * step + starting col
	offset int
	// analogous to dense step
	step int
}
*/

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

type serializableSparseMatrix struct {
	Rows   int
	Cols   int
	Keys   []int
	Values []float64
	Step   int
}

// map[string]float64 to map[int]float64
func kvToMap(keys []int, values []float64) (map[int]float64, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("unequal keys and values")
	}

	output := make(map[int]float64)
	for i, _ := range keys {
		output[keys[i]] = values[i]
	}

	return output, nil
}

// map[int]float64 to map[string]float64
func mapToKv(input map[int]float64) ([]int, []float64) {
	ints := make([]int, len(input))
	floats := make([]float64, len(input))

	i := 0
	for k, v := range input {
		ints[i] = k
		floats[i] = v
		i++
	}

	return ints, floats
}

func (A *SparseMatrix) MarshalJSON() ([]byte, error) {
	ints, floats := mapToKv(A.elements)
	return json.Marshal(&serializableSparseMatrix{
		Rows:   A.Rows(),
		Cols:   A.Cols(),
		Keys:   ints,
		Values: floats,
		Step:   A.step,
	})
}

func (A *SparseMatrix) UnmarshalJSON(b []byte) error {
	tmp := &serializableSparseMatrix{}
	err := json.Unmarshal(b, tmp)
	if err != nil {
		return err
	}

	elements, err := kvToMap(tmp.Keys, tmp.Values)
	if err != nil {
		return err
	}

	A.rows = tmp.Rows
	A.cols = tmp.Cols
	A.elements = elements
	A.step = tmp.Step

	return nil
}

func (A *SparseMatrix) GobEncode() ([]byte, error) {
	var b bytes.Buffer
	ints, floats := mapToKv(A.elements)
	tmp := &serializableSparseMatrix{
		Rows:   A.Rows(),
		Cols:   A.Cols(),
		Keys:   ints,
		Values: floats,
		Step:   A.step,
	}

	err := gob.NewEncoder(&b).Encode(tmp)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (A *SparseMatrix) GobDecode(b []byte) error {
	buf := bytes.NewBuffer(b)
	tmp := &serializableSparseMatrix{}
	err := gob.NewDecoder(buf).Decode(tmp)
	if err != nil {
		return err
	}

	elements, err := kvToMap(tmp.Keys, tmp.Values)
	if err != nil {
		return err
	}

	A.rows = tmp.Rows
	A.cols = tmp.Cols
	A.elements = elements
	A.step = tmp.Step

	return nil
}
