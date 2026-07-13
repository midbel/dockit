package grid

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type formulaView struct {
	view View
}

func FormulaView(view View) value.Value {
	return &formulaVliew{
		view: view,
	}
}

func (*formulaView) Type() string {
	return value.TypeArray
}

func (*formulaView) Kind() value.ValueKind {
	return value.KindArray
}

func (v *formulaView) String() string {
	return fmt.Sprintf("%s(%s)", value.TypeArray, v.view.Name())
}

func (v *formulaView) Unwrap() View {
	return v.inner
}

func (v *formulaView) Dimension() layout.Dimension {
	rg := v.inner.Bounds()
	dm := layout.Dimension{
		Lines:   int64(rg.Height()),
		Columns: int64(rg.Width()),
	}
	return dm
}

func (v *formulaView) At(row, col int) value.Value {
	pos := layout.Position{
		Line:   int64(row) + 1,
		Column: int64(col) + 1,
	}
	c, _ := v.inner.Cell(pos)
	return NewFormulaFromPosition(c.At())
}

func (a formulaView) Values() iter.Seq[value.Value] {
	it := func(yield func(value.Value) bool) {
		for _, rs := range a.view.Rows() {
			for _, v := range rs {
				ok := yield(v)
				if !ok {
					return
				}
			}
		}
	}
	return it
}

type arrayView struct {
	inner View
}

func ArrayView(view View) value.Value {
	v := arrayView{
		inner: view,
	}
	return &v
}

func (*arrayView) Type() string {
	return value.TypeArray
}

func (*arrayView) Kind() value.ValueKind {
	return value.KindArray
}

func (v *arrayView) String() string {
	return fmt.Sprintf("%s(%s)", value.TypeArray, v.inner.Name())
}

func (v *arrayView) Unwrap() View {
	return v.inner
}

func (v *arrayView) Dimension() layout.Dimension {
	rg := v.inner.Bounds()
	dm := layout.Dimension{
		Lines:   int64(rg.Height()),
		Columns: int64(rg.Width()),
	}
	return dm
}

func (v *arrayView) At(row, col int) value.Value {
	pos := layout.Position{
		Line:   int64(row) + 1,
		Column: int64(col) + 1,
	}
	c, _ := v.inner.Cell(pos)
	return c.Value()
}

func (a arrayView) Values() iter.Seq[value.Value] {
	it := func(yield func(value.Value) bool) {
		for _, rs := range a.inner.Rows() {
			for _, v := range rs {
				ok := yield(v)
				if !ok {
					return
				}
			}
		}
	}
	return it
}

func (a arrayView) AsArray() value.ArrayValue {
	var data [][]value.Value
	for _, r := range a.inner.Rows() {
		tmp := make([]value.Value, 0, len(r))
		tmp = append(tmp, r...)
		data = append(data, tmp)
	}
	return value.NewArray(data)
}

func (a arrayView) Apply(do func(value.Value) value.Value) {
	bd := a.inner.Bounds()
	mv, ok := a.inner.(MutableView)
	if !ok {
		return
	}
	for pos := range bd.Positions() {
		v := do(a.At(int(pos.Line), int(pos.Column)))
		mv.SetValue(pos, v)
	}
}

func (a arrayView) ApplyArray(other value.Array, do func(value.Value, value.Value) value.Value) value.Value {
	arr := a.AsArray()
	if arr, ok := arr.(value.Array); ok {
		return arr.ApplyArray(other, do)
	}
	return value.NewArray(nil)
}

func (a arrayView) Cells() [][]Cell {
	return cellsFromView(a.inner)
}
