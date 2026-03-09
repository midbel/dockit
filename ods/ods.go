package ods

import (
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Cell struct {
	layout.Position

	raw     string
	parsed  value.ScalarValue
	formula value.Formula
}

func (c *Cell) At() layout.Position {
	return c.Position
}

func (c *Cell) Display() string {
	return c.raw
}

func (c *Cell) Value() value.ScalarValue {
	if c.parsed == nil {
		return value.Empty()
	}
	return c.parsed
}

func (c *Cell) Reload(ctx value.Context) error {
	return nil
}

type row struct {
	Line   int64
	Hidden bool
	Cells  []*Cell
}

func (r *row) Values() []value.ScalarValue {
	var ds []value.ScalarValue
	for _, c := range r.Cells {
		ds = append(ds, c.Value())
	}
	return ds
}

func (r *row) Sparse() bool {
	for i := range r.Cells {
		if i == 0 {
			continue
		}
		if r.Cells[i].Column-r.Cells[i-1].Column > 1 {
			return true
		}
	}
	return false
}

type Sheet struct {
	Label  string
	Active bool
	Size   layout.Dimension

	rows  []*row
	cells map[layout.Position]*Cell
}

func NewSheet(name string) *Sheet {
	sh := Sheet{
		Label:  name,
		Active: false,
		cells:  make(map[layout.Position]*Cell),
	}
	return &sh
}

func (s *Sheet) Name() string {
	return s.Label
}

type File struct {
	names  map[string]int
	sheets []*Sheet
}

func NewFile() *File {
	return &File{
		names: make(map[string]int),
	}
}

func Open(file string) (*File, error) {
	rs, err := readFile(file)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	book, err := rs.ReadFile()
	if err != nil {
		return nil, err
	}
	return book, nil
}
