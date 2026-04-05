package types

import (
	"errors"
	"os"

	"github.com/midbel/dockit/grid"
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
	err := c.file.Sync()
	if errors.Is(err, grid.ErrSupported) {
		err = nil
	}
	return err
}

func (c *File) Active() (value.Value, error) {
	return c.Sheet("")
}

func (c *File) Sheet(ident string) (value.Value, error) {
	var (
		sh  grid.View
		err error
	)
	if ident == "" {
		sh, err = c.file.ActiveSheet()
	} else {
		sh, err = c.file.Sheet(ident)
	}
	if err != nil {
		return value.ErrNA, nil
	}
	return newView(sh, grid.FileContext(c.file), c.ro), nil
}

func (c *File) File() grid.File {
	return c.file
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
		v, _ := c.Sheet("")
		return v
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
