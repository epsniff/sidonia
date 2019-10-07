package index

import (
	"time"

	"github.com/araddon/qlbridge/value"
)

// TODO replace use of the Doc struct below with this interface
// replace the doc look in segment.go to use the Analizier
type Document interface {
	ID() string
	Get(key string) (value.Value, bool)
	Row() map[string]value.Value
	Ts() time.Time
}

type doc struct {
	externalId string
	fieldvals  map[string]value.Value
	ts         time.Time
}

func NewDocument(externalId string, fieldvals map[string]value.Value, ts time.Time) Document {
	return &doc{externalId, fieldvals, ts}
}

func (d *doc) ID() string {
	return d.externalId
}
func (d *doc) Get(key string) (value.Value, bool) {
	v, ok := d.fieldvals[key]
	return v, ok
}
func (d *doc) Row() map[string]value.Value {
	return d.fieldvals
}
func (d *doc) Ts() time.Time {
	return d.ts
}
