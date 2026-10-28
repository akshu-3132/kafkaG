package storage

type Index struct {
	Entries []int64
}

func NewIndex() *Index {
	return &Index{
		Entries: make([]int64, 0),
	}
}

func (idx *Index) AddEntry(offset int64) int64 {
	logicalOffset := int64(len(idx.Entries))
	idx.Entries = append(idx.Entries, offset)
	return logicalOffset
}

func (idx *Index) Get(logicalOffset int64) (int64, bool) {
	if logicalOffset < 0 || logicalOffset >= int64(len(idx.Entries)) {
		return 0, false
	}
	return idx.Entries[logicalOffset], true
}
