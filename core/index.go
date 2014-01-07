package core

const (
	IndexType = iota + 1
	UniqueType
)

// database index
type Index struct {
	Name string
	Type int
	Cols []string
}

// add columns which will be composite index
func (index *Index) AddColumn(cols ...string) {
	for _, col := range cols {
		index.Cols = append(index.Cols, col)
	}
}

// new an index
func NewIndex(name string, indexType int) *Index {
	return &Index{name, indexType, make([]string, 0)}
}
