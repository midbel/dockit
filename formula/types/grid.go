package types

import (
	"fmt"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type fileValue struct {
	file grid.File
}

func NewFileValue(file grid.File) value.Value {
	return &fileValue{
		file: file,
	}
}

func (*fileValue) Kind() value.ValueKind {
	return value.KindObject
}

func (*fileValue) String() string {
	return "workbook"
}

func (c *fileValue) Get(ident string) (value.Value, error) {
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
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
}

type viewValue struct {
	view grid.View
}

func NewViewValue(view grid.View) value.Value {
	return &viewValue{
		view: view,
	}
}

func (*viewValue) Kind() value.ValueKind {
	return value.KindObject
}

func (c *viewValue) String() string {
	return c.view.Name()
}

func (c *viewValue) Get(ident string) (value.Value, error) {
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
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
}

type rangeValue struct {
	rg *layout.Range
}

func (*rangeValue) Kind() value.ValueKind {
	return value.KindObject
}

func (v *rangeValue) String() string {
	return v.rg.String()
}

func (v *rangeValue) Get(name string) (value.ScalarValue, error) {
	return nil, nil
}

// type lambdaValue struct {
// 	expr Expr
// }

// func (*lambdaValue) Kind() value.ValueKind {
// 	return value.KindFunction
// }

// func (*lambdaValue) String() string {
// 	return "<formula>"
// }

// func (v *lambdaValue) Call(args []value.Arg, ctx value.Context) (value.Value, error) {
// 	return Eval(v.expr, ctx)
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
