package index

// IndexableField
type IndexableField struct {
	InternalDocId uint32 // TODO use uint16 instead?  And limit the size of the segment to 65k docs?
	FieldID       uint32
	FieldName     string
	Terms         Terms
}
type IndexableFields map[uint32]*IndexableField

type Term struct {
	Term   string
	TermID uint32
}
type Terms []*Term

func (p Terms) Len() int           { return len(p) }
func (p Terms) Less(i, j int) bool { return p[i].Term < p[j].Term }
func (p Terms) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
