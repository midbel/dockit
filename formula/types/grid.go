package types

import (
	"fmt"
	"os"
	"slices"

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

func (c *File) Get(ident string) (value.Value, error) {
	switch ident {
	case "names":
		return value.Float(0), nil
	case "sheets":
		x := c.file.Sheets()
		return value.Float(float64(len(x))), nil
	case "readonly":
		return value.Boolean(c.ro), nil
	case "protected":
		return value.Boolean(false), nil
	case "active":
		sh, err := c.file.ActiveSheet()
		if err != nil {
			return value.ErrValue, nil
		}
		return newView(sh, c.ro), nil
	default:
		v, err := c.Sheet(ident)
		if err == nil {
			return v, nil
		}
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

func (v *View) Type() string {
	if t, ok := v.view.(interface{ Type() string }); ok {
		return t.Type()
	}
	return "view"
}

func (*View) Kind() value.ValueKind {
	return value.KindObject
}

func (c *View) String() string {
	return c.view.Name()
}

func (c *View) FilterView(predicate value.Predicate) {
	c.view = grid.FilterView(c.view, predicate)
}

func (c *View) ProjectView(sel layout.Selection) {
	c.view = grid.NewProjectView(c.view, sel)
}

func (c *View) BoundedView(rg *layout.Range) {
	c.view = grid.NewBoundedView(c.view, rg)
}

func (c *View) Inspect() *InspectValue {
	var (
		iv = InspectView()
		bs = c.view.Bounds()
	)
	iv.Set("name", value.Text(c.view.Name()))
	iv.Set("rows", value.Float(bs.Height()))
	iv.Set("cols", value.Float(bs.Width()))
	iv.Set("type", value.Text(c.Type()))

	return iv
}

func (c *View) Get(ident string) (value.Value, error) {
	switch ident {
	case "name":
		return value.Text(c.view.Name()), nil
	case "lines":
		rg := c.view.Bounds()
		lines := rg.Ends.Line - rg.Starts.Line
		return value.Float(float64(lines)), nil
	case "columns":
		rg := c.view.Bounds()
		lines := rg.Ends.Column - rg.Starts.Column
		return value.Float(float64(lines)), nil
	case "cells":
		var count int
		for x := range c.view.Rows() {
			count += len(x)
		}
		return value.Float(float64(count)), nil
	case "empty":
		return value.Float(float64(0)), nil
	case "readonly":
		return value.Boolean(c.ro), nil
	case "protected":
		var locked bool
		if k, ok := c.view.(interface{ IsLock() bool }); ok {
			locked = k.IsLock()
		}
		return value.Boolean(locked), nil
	case "active":
		return value.Boolean(false), nil
	case "index":
		return value.Float(float64(0)), nil
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

func (c *View) AsArray() value.ArrayValue {
	var data [][]value.ScalarValue
	for r := range c.view.Rows() {
		data = append(data, slices.Clone(r))
	}
	return value.NewArray(data)
}

type Range struct {
	rg *layout.Range
}

func NewRangeValue(start, end layout.Position) value.Value {
	rg := layout.NewRange(start, end)
	return &Range{
		rg: rg.Normalize(),
	}
}

func (v *Range) Type() string {
	return "range"
}

func (*Range) Kind() value.ValueKind {
	return value.KindObject
}

func (v *Range) String() string {
	return v.rg.String()
}

func (v *Range) Target() string {
	return v.rg.Starts.Sheet
}

func (v *Range) Range() *layout.Range {
	return v.rg
}

func (v *Range) Get(ident string) (value.ScalarValue, error) {
	switch ident {
	case "lines":
		return value.Float(v.rg.Height()), nil
	case "columns":
		return value.Float(v.rg.Width()), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, value.ErrProp)
	}
	return nil, nil
}

func (v *Range) Collect(view grid.View) (value.Value, error) {
	var (
		width  = v.rg.Width()
		height = v.rg.Height()
		data   = make([][]value.ScalarValue, height)
		col    int64
		row    int64
	)
	for i := range data {
		data[i] = make([]value.ScalarValue, width)
	}
	for pos := range v.rg.Positions() {
		cell, err := view.Cell(pos)
		if err != nil {
			return nil, err
		}
		val := cell.Value()
		if val == nil {
			val = value.Empty()
		}
		data[row][col] = val

		col++
		if col == width {
			row++
			col = 0
		}
	}
	return value.NewArray(data), nil
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
