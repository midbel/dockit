package runtime

import (
	"fmt"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Cell struct {
	grid.Cell
	view *View
}

type cellsArray struct {
	cells [][]grid.Cell
}

func NewCellsArray(cells [][]grid.Cell) value.ArrayValue {
	return &cellsArray{
		cells: cells,
	}
}

func (*cellsArray) Type() string {
	return value.TypeArray
}

func (*cellsArray) Kind() value.ValueKind {
	return value.KindArray
}

func (v *cellsArray) String() string {
	return fmt.Sprintf("%s(cells)", value.TypeArray)
}

func (v *cellsArray) Dimension() layout.Dimension {
	var (
		rows = len(v.cells)
		cols int
	)
	for _, cs := range v.cells {
		cols = max(cols, len(cs))
	}
	dm := layout.Dimension{
		Lines:   int64(rows),
		Columns: int64(cols),
	}
	return dm
}

func (v *cellsArray) At(row, col int) value.Value {
	if row < 0 || row >= len(v.cells) {
		return value.Empty()
	}
	cs := v.cells[row]
	if col < 0 || row >= len(cs) {
		return value.Empty()
	}
	return grid.NewFormulaFromPosition(cs[col].At())
}
