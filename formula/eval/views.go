package eval

import (
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
	"github.com/midbel/dockit/layout"
)

type scalarView struct {
	scalar value.ScalarValue
}

func NewScalarView(scalar value.ScalarValue) grid.View {
	return scalarView{
		scalar: scalar,
	}
}

func (scalarView) Name() string {
	return "scalar"
}

func (a scalarView) Bounds() *layout.Range {
	return nil
}

func (a scalarView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	return nil
}

func (a scalarView) Cell(pos layout.Position) (grid.Cell, error) {
	return nil, nil
}

func (scalarView) Sync(value.Context) error {
	return grid.ErrSupported
}


type arrayView struct {
	array value.ArrayValue
}

func NewArrayView(array value.ArrayValue) grid.View {
	return arrayView{
		array: array,
	}
}

func (arrayView) Name() string {
	return "array"
}

func (a arrayView) Bounds() *layout.Range {
	return nil
}

func (a arrayView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	return nil
}

func (a arrayView) Cell(pos layout.Position) (grid.Cell, error) {
	return nil, nil
}

func (arrayView) Sync(value.Context) error {
	return grid.ErrSupported
}