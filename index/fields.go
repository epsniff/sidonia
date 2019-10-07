package index

// IndexableField
type IndexableField struct {
	FieldID   uint32
	FieldName string

	// termID to postings list DocIds
	terminIdInt  uint32 // Question use uint16 instead?  And limit the size of the segment to 65k terms per field?
	termToTermID map[string]uint32

	Terms Terms
}

func NewIndexableField(field string, fieldID uint32) *IndexableField {
	return &IndexableField{
		FieldID:      fieldID,
		FieldName:    field,
		termToTermID: map[string]uint32{},
	}
}

type IndexableFields map[uint32]*IndexableField

type Term struct {
	Term          string
	TermID        uint32
	InternalDocId uint32 // TODO use uint16 instead?  And limit the size of the segment to 65k docs?
}
type Terms []*Term

func (p Terms) Len() int           { return len(p) }
func (p Terms) Less(i, j int) bool { return p[i].Term < p[j].Term }
func (p Terms) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
