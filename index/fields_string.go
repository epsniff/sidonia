package index

import (
	"github.com/araddon/qlbridge/value"
)

type stringVal struct {
	val string
}

func NewStringVal(val string) *stringVal {
	return &stringVal{val}
}

// Is this a nil/empty?
// empty string counts as nil, empty slices/maps, nil structs.
func (s *stringVal) Nil() bool {
	return s.val == ""
}

// Is this an error, or unable to evaluate from Vm?
func (s *stringVal) Err() bool {
	return false
}
func (s *stringVal) Value() interface{} {
	return s.val
}
func (s *stringVal) ToString() string {
	return s.val
}
func (s *stringVal) Type() value.ValueType {
	return value.StringType
}
