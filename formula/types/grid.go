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
		ro: readonly,
	}
}

func (*File) Kind() value.ValueKind {
	return value.KindObject
}

func (*File) String() string {
	return "workbook"
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

func (c *File) Get(ident string) (value.Value, error) {
	switch ident {
	case "names":
		return Float(0), nil
	case "sheets":
		x := c.file.Sheets()
		return Float(float64(len(x))), nil
	case "readonly":
		return Boolean(c.ro), nil
	case "protected":
		return Boolean(false), nil
	case "active":
		sh, err := c.file.ActiveSheet()
		if err != nil {
			return ErrValue, nil
		}
		return newView(sh, c.ro), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, value.ErrProp)
	}
}

type View struct {
	view grid.View
	ro   bool
}

func NewViewValue(view grid.View) value.Value {
	return newView(view, false)
}

func newView(view grid.View, ro bool) value.Value {
	return &View{
		view: view,
		ro:   ro,
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
	case "readonly":
		return Boolean(c.ro), nil
	case "protected":
		var locked bool
		if k, ok := c.view.(interface{ IsLock() bool }); ok {
			locked = k.IsLock()
		}
		return Boolean(locked), nil
	case "active":
		return Boolean(false), nil
	case "index":
		return Float(float64(0)), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, value.ErrProp)
	}
}

func (c *View) View() grid.View {
	if c.ro {
		return grid.ReadOnly(c.view)
	}
	return c.view
}

func (c *View) Mutable() (grid.MutableView, error) {
	if c.ro {
		return nil, fmt.Errorf("view is not mutable")
	}
	mv, ok := c.view.(grid.MutableView)
	if !ok {
		return nil, fmt.Errorf("view is not mutable")
	}
	return mv, nil
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
