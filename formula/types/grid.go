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
}

func NewFileValue(file grid.File) value.Value {
	return &File{
		file: file,
	}
}

func (*File) Kind() value.ValueKind {
	return value.KindObject
}

func (*File) String() string {
	return "workbook"
}

func (c *File) Sheet(ident string) (value.Value, error) {
	v, err := c.file.Sheet(ident)
	if err != nil {
		return nil, err
	}
	return NewViewValue(v), nil
}

func (c *File) Get(ident string) (value.Value, error) {
	switch ident {
	case "sheets":
		x := c.file.Sheets()
		return Float(float64(len(x))), nil
	case "protected":
		return Boolean(false), nil
	case "active":
		sh, err := c.file.ActiveSheet()
		if err != nil {
			return ErrValue, nil
		}
		return NewViewValue(sh), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, value.ErrProp)
	}
}

type View struct {
	view grid.View
}

func NewViewValue(view grid.View) value.Value {
	return &View{
		view: view,
	}
}

func (*View) Kind() value.ValueKind {
	return value.KindObject
}

func (c *View) String() string {
	return c.view.Name()
}

func (c *View) Get(ident string) (value.Value, error) {
	switch ident {
	case "name":
		return Text(c.view.Name()), nil
	case "lines":
		rg := c.view.Bounds()
		lines := rg.Ends.Line - rg.Starts.Line
		return Float(float64(lines)), nil
	case "columns":
		rg := c.view.Bounds()
		lines := rg.Ends.Column - rg.Starts.Column
		return Float(float64(lines)), nil
	case "cells":
		var count int
		for x := range c.view.Rows() {
			count += len(x)
		}
		return Float(float64(count)), nil
	case "empty":
		return Float(float64(0)), nil
	case "protected":
		var locked bool
		if k, ok := c.view.(interface{ IsLock() bool }); ok {
			locked = k.IsLock()
		}
		return Boolean(locked), nil
	case "readonly":
		return Boolean(false), nil
	case "active":
		return Boolean(false), nil
	case "index":
		return Float(float64(0)), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, value.ErrProp)
	}
}

type Range struct {
	rg *layout.Range
}

func (*Range) Kind() value.ValueKind {
	return value.KindObject
}

func (v *Range) String() string {
	return v.rg.String()
}

func (v *Range) Get(name string) (value.ScalarValue, error) {
	return nil, nil
}

// type Lambda struct {
// 	expr eval.Expr
// }

// func (*Lambda) Kind() value.ValueKind {
// 	return value.KindFunction
// }

// func (*Lambda) String() string {
// 	return "<formula>"
// }

// func (v *Lambda) Call(args []value.Arg, ctx value.Context) (value.Value, error) {
// 	return eval.Eval(v.expr, ctx)
// }

type envValue struct{}

func (envValue) Kind() value.ValueKind {
	return value.KindObject
}

func (envValue) String() string {
	return "env"
}

func (v envValue) Get(name string) (value.ScalarValue, error) {
	str := os.Getenv(name)
	return Text(str), nil
}
