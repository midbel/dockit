package types

import (
	"fmt"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

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
