package index

// IndexableField
type IndexableField struct {
	InternalDocId uint32 // TODO use uint16 instead?  And limit the size of the segment to 65k docs?
	Term          string
	TermID        uint32
}
type IndexableFields []*IndexableField

func (p IndexableFields) Len() int           { return len(p) }
func (p IndexableFields) Less(i, j int) bool { return p[i].Term < p[j].Term }
func (p IndexableFields) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
