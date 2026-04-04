package types

import (
	"fmt"
	"slices"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type View struct {
	view grid.View
	ro   bool
}

func NewViewValue(view grid.View) value.Value {
	return newView(view, false)
}

func newView(view grid.View, ro bool) value.Value {
	return createView(view, ro)
}

func createView(view grid.View, ro bool) *View {
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

// func (c *View) FilterView(predicate value.Predicate) *View {
// 	view := grid.FilterView(c.view, predicate)
// 	return createView(view, false)
// }

func (c *View) ProjectView(sel layout.Selection) *View {
	view := grid.NewProjectView(c.view, sel)
	return createView(view, false)
}

func (c *View) BoundedView(rg *layout.Range) *View {
	view := grid.NewBoundedView(c.view, rg)
	return createView(view, false)
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

func (c *View) At(pos layout.Position) value.Value {
	cell, err := c.view.Cell(pos)
	if err != nil {
		return value.ErrNA
	}
	return cell.Value()
}

func (c *View) Range(start, end layout.Position) value.Value {
	rg := layout.NewRange(start, end)
	return grid.ArrayView(grid.NewBoundedView(c.view, rg))
}

func (c *View) Get(ident string) value.Value {
	switch ident {
	case "name":
		return value.Text(c.view.Name())
	case "lines":
		rg := c.view.Bounds()
		lines := rg.Ends.Line - rg.Starts.Line
		return value.Float(float64(lines))
	case "columns":
		rg := c.view.Bounds()
		lines := rg.Ends.Column - rg.Starts.Column
		return value.Float(float64(lines))
	case "cells":
		var count int
		for _, x := range c.view.Rows() {
			count += len(x)
		}
		return value.Float(float64(count))
	case "empty":
		return value.Float(float64(0))
	case "readonly":
		return value.Boolean(c.ro)
	case "protected":
		var locked bool
		if k, ok := c.view.(interface{ IsLock() bool }); ok {
			locked = k.IsLock()
		}
		return value.Boolean(locked)
	case "active":
		return value.Boolean(false)
	case "index":
		return value.Float(float64(0))
	default:
		return value.ErrName
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
	for _, r := range c.view.Rows() {
		data = append(data, slices.Clone(r))
	}
	return value.NewArray(data)
}
