package grid

import (
	"errors"
	"fmt"
	"iter"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrFile      = errors.New("invalid spreadsheet")
	ErrLock      = errors.New("spreadsheet locked")
	ErrSupported = errors.New("operation not supported")
	ErrFound     = errors.New("not found")
)

func NoCell(pos layout.Position) error {
	return fmt.Errorf("%s no cell at given position", pos)
}

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

type CellType int8

const (
	TypeSharedString CellType = 1 << iota
	TypeString
	TypeNumber
	TypeDate
	TypeBool
	TypeFormula
)

type Cell interface {
	At() layout.Position
	Value() value.ScalarValue
	Reload(value.Context) error
	// Type() CellType
}

type Encoder interface {
	EncodeSheet(View) error
}

type Callable interface {
	Call(value.Context) (value.Value, error)
}

type View interface {
	Name() string
	Bounds() *layout.Range
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
	Cell(layout.Position) (Cell, error)

	Reload(value.Context) error
}

type MutableView interface {
	View

	SetValue(layout.Position, value.ScalarValue) error
	SetFormula(layout.Position, value.Formula) error

	ClearCell(layout.Position) error
	ClearValue(layout.Position) error
	ClearFormula(layout.Position) error
	ClearRange(*layout.Range) error

	AppendRow([]value.ScalarValue) error
	InsertRow(int64, []value.ScalarValue) error
	DeleteRow(int64) error
}

type ViewInfo struct {
	Name      string
	Active    bool
	Hidden    bool
	Size      layout.Dimension
	Protected bool
}

type File interface {
	Infos() []ViewInfo
	ActiveSheet() (View, error)
	Sheet(string) (View, error)
	Sheets() []View

	Reload() error

	// Merge(File) error
	Rename(string, string) error
	Copy(string, string) error
	Remove(string) error
}

type filteredView struct {
	sheet View
}

func FilterView(view View) View {
	return &filteredView{
		sheet: view,
	}
}

func (v *filteredView) Name() string {
	return v.sheet.Name()
}

func (v *filteredView) Bounds() *layout.Range {
	return nil
}

func (v *filteredView) Rows() iter.Seq[[]value.ScalarValue] {
	return nil
}

func (v *filteredView) Encode(encoder Encoder) error {
	return encoder.EncodeSheet(v.sheet)
}

func (v *filteredView) Cell(layout.Position) (Cell, error) {
	return nil, nil
}

func (v *filteredView) Reload(ctx value.Context) error {
	return v.sheet.Reload(ctx)
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

func (v *projectedView) Reload(ctx value.Context) error {
	return v.sheet.Reload(ctx)
}

func (v *projectedView) Bounds() *layout.Range {
	rg := v.sheet.Bounds()

	start := layout.Position{
		Line:   1,
		Column: 1,
	}
	if rg.Width() == 0 && rg.Height() == 0 {
		return layout.NewRange(start, start)
	}
	end := layout.Position{
		Line:   rg.Height(),
		Column: max(1, int64(len(v.columns))),
	}
	return layout.NewRange(start, end)
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

func (v *boundedView) Reload(ctx value.Context) error {
	return v.sheet.Reload(ctx)
}

func (v *boundedView) Cell(pos layout.Position) (Cell, error) {
	if !v.part.Contains(pos) {
		return nil, fmt.Errorf("position outside view range")
	}
	return v.sheet.Cell(pos)
}

func (v *boundedView) Bounds() *layout.Range {
	return v.part.Range()
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
