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
	Name() string
	Bounds() *layout.Range
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
	Cell(layout.Position) (Cell, error)
}

type MutableView interface {
	View
	AppendRow([]value.ScalarValue) error
	DeleteRow(int) error
	InsertRow(int, []value.ScalarValue) error
}

type File interface {
	Merge(File) error

	ActiveSheet() View
	Sheets() ([]View, error)
	Copy(View) error
	Move(View) error
}

type projectedView struct {
	sheet   View
	columns []int64
	mapping map[int64]int64
}

func NewProjectView(view View, sel layout.Selection) View {
	v := projectedView{
		sheet:   view,
		columns: sel.Indices(view.Bounds()),
		mapping: make(map[int64]int64),
	}
	for i, c := range v.columns {
		v.mapping[c] = int64(i)
	}
	return &v
}

func (v *projectedView) Name() string {
	return v.sheet.Name()
}

func (v *projectedView) Bounds() *layout.Range {
	return v.sheet.Bounds()
}

func (v *projectedView) Cell(pos layout.Position) (Cell, error) {
	if pos.Column < 0 || pos.Column > int64(len(v.columns)) {
		return nil, nil
	}
	mod := layout.Position{
		Column: v.columns[pos.Column],
		Line:   pos.Line,
	}
	return v.sheet.Cell(mod)
}

func (v *projectedView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		out := make([]value.ScalarValue, len(v.columns))
		for row := range v.sheet.Rows() {
			for i, col := range v.columns {
				if int(col) < len(row) {
					out[i] = row[col]
				}
			}
			if !yield(out) {
				return
			}
		}
	}
	return it
}

func (v *projectedView) Encode(encoder Encoder) error {
	return encoder.EncodeSheet(v)
}

type boundedView struct {
	sheet View
	part  *layout.Range
}

func NewBoundedView(view View, rg *layout.Range) View {
	v := boundedView{
		sheet: view,
		part:  rg.Normalize(),
	}
	return &v
}

func (v *boundedView) Name() string {
	return v.sheet.Name()
}

func (v *boundedView) Cell(pos layout.Position) (Cell, error) {
	if !v.part.Contains(pos) {
		return nil, fmt.Errorf("position outside view range")
	}
	return v.sheet.Cell(pos)
}

func (v *boundedView) Bounds() *layout.Range {
	return v.part
}

func (v *boundedView) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		var (
			width = v.part.Ends.Column - v.part.Starts.Column + 1
			data  = make([]value.ScalarValue, width)
		)
		for row := v.part.Starts.Line; row <= v.part.Ends.Line; row++ {
			for col, ix := v.part.Starts.Column, 0; col <= v.part.Ends.Column; col++ {
				p := layout.Position{
					Line:   row,
					Column: col,
				}
				c, err := v.sheet.Cell(p)
				if err == nil {
					data[ix] = c.Value()
				} else {
					data[ix] = formula.Blank{}
				}
				ix++
			}
			if !yield(data) {
				break
			}
		}
	}
	return it
}

func (v *boundedView) Encode(e Encoder) error {
	return e.EncodeSheet(v)
}
