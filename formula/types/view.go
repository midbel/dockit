package types

import (
	"slices"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type View struct {
	view grid.View
	ro   bool
	ctx  value.Context
}

func NewViewValue(view grid.View) value.Value {
	return newView(view, nil, false)
}

func newView(view grid.View, ctx value.Context, ro bool) value.Value {
	return createView(view, ctx, ro)
}

func createView(view grid.View, ctx value.Context, ro bool) *View {
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

func (c *View) Sync() error {
	var (
		sub = grid.SheetContext(c.view)
		ctx value.Context
	)
	if c.ctx != nil {
		ctx = grid.EnclosedContext(ctx, sub)
	} else {
		ctx = sub
	}
	return c.view.Sync(ctx)
}

// func (c *View) FilterView(predicate value.Predicate) *View {
// 	view := grid.FilterView(c.view, predicate)
// 	return createView(view, false)
// }

func (c *View) ProjectView(sel layout.Selection) *View {
	view := grid.NewProjectView(c.view, sel)
	return createView(view, c.ctx, false)
}

func (c *View) BoundedView(rg *layout.Range) *View {
	view := grid.NewBoundedView(c.view, rg)
	return createView(view, c.ctx, false)
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
	if err := c.syncIfNeeded(cell); err != nil {
		return value.ErrValue
	}
	return cell.Value()
}

func (c *View) FormulaAt(pos layout.Position) value.Formula {
	cell, err := c.view.Cell(pos)
	if err != nil {
		return nil
	}
	return cell.Formula()
}

func (c *View) SetAt(pos layout.Position, val value.Value) error {
	if c.ro {
		return ErrReadOnly
	}
	mv, ok := c.view.(grid.MutableView)
	if !ok {
		return ErrReadOnly
	}
	if f, ok := val.(parse.Deferred); ok {
		return mv.SetFormula(pos, grid.NewFormula(f.Expr()))
	}
	scalar, ok := val.(value.ScalarValue)
	if !ok {
		return ErrType
	}
	return mv.SetValue(pos, scalar)
}

func (c *View) Range(start, end layout.Position) value.Value {
	rg := layout.NewRange(start, end)
	for pos := range rg.Positions() {
		cell, err := c.view.Cell(pos)
		if err != nil {
			continue
		}
		if err := c.syncIfNeeded(cell); err != nil {
			return value.ErrValue
		}
	}
	return grid.ArrayView(grid.NewBoundedView(c.view, rg))
}

func (c *View) SetRange(start, end layout.Position, val value.Value) error {
	if c.ro {
		return ErrReadOnly
	}
	rg := layout.NewRange(start, end)
	switch v := val.(type) {
	case parse.Deferred:
		fm := grid.NewFormula(v.Expr())
		for pos := range rg.Positions() {
			if err := c.SetAt(pos, fm); err != nil {
				return err
			}
		}
	case value.ScalarValue:
		for pos := range rg.Positions() {
			if err := c.SetAt(pos, val); err != nil {
				return err
			}
		}
	case value.ArrayValue:
		return c.setArray(v, rg)
	default:
		return ErrType
	}
	return nil
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

func (c *View) syncIfNeeded(cell grid.Cell) error {
	if !cell.Dirty() {
		return nil
	}
	if s, ok := cell.(interface{ Sync(value.Context) error }); ok {
		return s.Sync(c.getContext())
	}
	return nil
}

func (c *View) getContext() value.Context {
	var (
		ctx  value.Context
		curr = grid.SheetContext(c.view)
	)
	if c.ctx != nil {
		ctx = grid.EnclosedContext(c.ctx, curr)
	} else {
		ctx = curr
	}
	return ctx
}

func (c *View) AsArray() value.ArrayValue {
	var data [][]value.ScalarValue
	for _, r := range c.view.Rows() {
		data = append(data, slices.Clone(r))
	}
	return value.NewArray(data)
}

func (c *View) setArray(arr value.ArrayValue, rg *layout.Range) error {
	mode, err := getBroadcastMode(rg, arr)
	if err != nil {
		return err
	}
	var (
		index int
		row   int
		col   int
		dim   = arr.Dimension()
	)
	for pos := range rg.Positions() {
		var val value.ScalarValue
		switch mode {
		case broadcastExact:
			val = arr.At(row, col)
		case broadcastRow:
			val = arr.At(0, col)
		case broadcastCol:
			val = arr.At(row, 0)
		case broadcastScalar:
			val = arr.At(0, 0)
		case broadcastFlat:
			r := index / int(dim.Lines)
			c := index % int(dim.Columns)
			val = arr.At(r, c)
			index++
		default:
			continue
		}
		if err := c.SetAt(pos, val); err != nil {
			return err
		}
		col++
		if col == int(rg.Width()) {
			col = 0
			row++
		}
	}
	return nil
}
