package eval

import (
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
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
	var (
		start = layout.NewPosition(1, 1)
		end   = layout.NewPosition(1, 1)
	)
	return layout.NewRange(start, end)
}

func (a scalarView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		yield(1, slx.One(a.scalar))
	}
	return it
}

func (a scalarView) Cell(pos layout.Position) (grid.Cell, error) {
	if pos.Line == 1 && pos.Column == 1 {
		return grid.Single(a.scalar, pos), nil
	}
	return grid.Empty(pos), nil
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
	var (
		dim   = a.array.Dimension()
		start = layout.NewPosition(1, 1)
		end   = layout.NewPosition(dim.Lines, dim.Columns)
	)
	return layout.NewRange(start, end)
}

func (a arrayView) AsArray() value.ArrayValue {
	return a.array
}

func (a arrayView) Rows() iter.Seq2[int64, []value.ScalarValue] {
	it := func(yield func(int64, []value.ScalarValue) bool) {
		dim := a.array.Dimension()
		for i := int64(0); i < dim.Lines; i++ {
			var out []value.ScalarValue
			for j := int64(0); j < dim.Columns; j++ {
				x := a.array.At(int(i), int(j))
				out = append(out, x)
			}
			if !yield(i+1, out) {
				return
			}
		}
	}
	return it
}

func (a arrayView) Cell(pos layout.Position) (grid.Cell, error) {
	x := a.array.At(int(pos.Line-1), int(pos.Column-1))
	return grid.Single(x, pos), nil
}

func (arrayView) Sync(value.Context) error {
	return grid.ErrSupported
}
