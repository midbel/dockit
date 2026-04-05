package types

import (
	"fmt"
	"os"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type File struct {
	file grid.File
	ro   bool
}

func NewFileValue(file grid.File, readonly bool) value.Value {
	return &File{
		file: file,
		ro:   readonly,
	}
}

func (*File) Type() string {
	return "file"
}

func (*File) Kind() value.ValueKind {
	return value.KindObject
}

func (f *File) Inspect() *InspectValue {
	var (
		iv = InspectFile()
		sz = len(f.file.Sheets())
	)
	iv.Set("sheets", value.Float(sz))
	if a, err := f.file.ActiveSheet(); err == nil {
		iv.Set("active", value.Text(a.Name()))
	}
	return iv
}

func (*File) String() string {
	return "workbook"
}

func (c *File) Sync() error {
	return c.file.Sync()
}

func (c *File) Active() (value.Value, error) {
	v, err := c.file.ActiveSheet()
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, fmt.Errorf("no active view")
	}
	return newView(v, c.ro), nil
}

func (c *File) Sheet(ident string) (value.Value, error) {
	v, err := c.file.Sheet(ident)
	if err != nil {
		return nil, err
	}
	return newView(v, c.ro), nil
}

func (c *File) File() grid.File {
	return c.file
}

func (c *File) At(pos layout.Position) value.Value {
	v, err := c.file.Sheet(pos.Sheet)
	if err != nil {
		return value.ErrNA
	}
	cell, err := v.Cell(pos)
	if err != nil {
		return value.ErrNA
	}
	return cell.Value()
}

func (c *File) Range(start, end layout.Position) value.Value {
	v, err := c.file.Sheet(start.Sheet)
	if err != nil {
		return value.ErrNA
	}
	rg := layout.NewRange(start, end)
	return grid.ArrayView(grid.NewBoundedView(v, rg))
}

func (c *File) Get(ident string) value.Value {
	switch ident {
	case "names":
		return value.Float(0)
	case "sheets":
		x := c.file.Sheets()
		return value.Float(float64(len(x)))
	case "readonly":
		return value.Boolean(c.ro)
	case "protected":
		return value.Boolean(false)
	case "active":
		sh, err := c.file.ActiveSheet()
		if err != nil {
			return value.ErrNA
		}
		return newView(sh, c.ro)
	default:
		v, err := c.Sheet(ident)
		if err == nil {
			return v
		}
		return value.ErrName
	}
}

type envValue struct{}

func (envValue) Kind() value.ValueKind {
	return value.KindObject
}

func (envValue) String() string {
	return "env"
}

func (v envValue) Get(name string) (value.ScalarValue, error) {
	str := os.Getenv(name)
	return value.Text(str), nil
}
