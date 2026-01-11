package grid

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type CopyMode int

func CopyModeFromString(str string) (CopyMode, error) {
	var mode CopyMode
	switch str {
	case "value":
		mode |= CopyValue
	case "formula":
		mode |= CopyFormula
	case "style":
		mode |= CopyStyle
	case "", "all":
		mode |= CopyAll
	default:
		return mode, fmt.Errorf("%s invalid value for copy mode", str)
	}
	return mode, nil
}

const (
	CopyValue = iota << 1
	CopyFormula
	CopyStyle
	CopyAll = CopyValue | CopyFormula | CopyStyle
)

type Row interface {
	Values() []value.ScalarValue
	Sparse() bool
}

type Cell interface {
	At() layout.Position
	Value() value.ScalarValue
	Reload(formula.Context) error
}

type Encoder interface {
	EncodeSheet(View) error
}

type View interface {
	Bounds() *layout.Range
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
	Cell(layout.Position) value.ScalarValue
	Cells() iter.Seq[Cell]
}

type MutableView interface {
	View
	AppendRow([]value.ScalarValue) error
	DeleteRow(int) error
	InsertRow(int, []value.ScalarValue) error
}
