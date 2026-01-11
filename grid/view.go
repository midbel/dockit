package grid

import (
	"fmt"
	"iter"
	"maps"
	"math"

	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type View interface {
	Name() string
	Bounds() *layout.Range
	Cell(layout.Position) (value.Value, error)
	// Cells() iter.Seq[value.Value]
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
}

type Cell struct {
	layout.Position
	Raw     string
	Parsed  value.ScalarValue
	Formula formula.Expr
	Dirty   bool
}

func (c *Cell) Reload(ctx formula.Context) error {
	if c.Formula == nil {
		return nil
	}
	res, err := formula.Eval(c.Formula, ctx)
	if err == nil {
		if !formula.IsScalar(res) {
			c.Parsed = formula.ErrValue
		} else {
			c.Parsed = res.(value.ScalarValue)
		}
		c.Raw = res.String()
	}
	return err
}

type Row struct {
	Line   int64
	Hidden bool
	Cells  []*Cell
}

func (r *Row) Data() []value.ScalarValue {
	var ds []value.ScalarValue
	for _, c := range r.Cells {
		ds = append(ds, c.Parsed)
	}
	return ds
}

func (r *Row) Values() []any {
	var list []any
	for i := range r.Cells {
		list = append(list, r.Cells[i].Parsed)
	}
	return list
}

func (r *Row) cloneCells() []*Cell {
	var cells []*Cell
	for i := range r.Cells {
		c := *r.Cells[i]
		cells = append(cells, &c)
	}
	return cells
}

type Sheet struct {
	name  string
	rows  []*Row
	cells map[layout.Position]*Cell
	Size  layout.Dimension
}

func Empty(name string) *Sheet {
	sh := Sheet{
		name:  name,
		cells: make(map[layout.Position]*Cell),
	}
	return &sh
}

func (s *Sheet) Name() string {
	return s.name
}

func (s *Sheet) Reload(ctx formula.Context) error {
	for _, c := range s.cells {
		if err := c.Reload(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sheet) View(rg *layout.Range) View {
	bd := s.Bounds()
	rg.Starts = rg.Starts.Update(bd.Starts)
	rg.Ends = rg.Ends.Update(bd.Ends)
	return newBoundedView(s, rg)
}

func (s *Sheet) Sub(start, end layout.Position) View {
	return s.View(layout.NewRange(start, end))
}

func (s *Sheet) Bounds() *layout.Range {
	var (
		minRow int64 = math.MaxInt64
		maxRow int64
		minCol int64 = math.MaxInt64
		maxCol int64
	)
	for c := range maps.Values(s.cells) {
		minRow = min(minRow, c.Line)
		maxRow = max(maxRow, c.Line)
		minCol = min(minCol, c.Column)
		maxCol = max(maxCol, c.Column)
	}
	var rg layout.Range
	if maxRow == 0 || maxCol == 0 {
		pos := layout.Position{
			Line:   1,
			Column: 1,
		}
		rg.Starts = pos
		rg.Ends = pos
	} else {
		rg.Starts = layout.Position{
			Line:   minRow,
			Column: minCol,
		}
		rg.Ends = layout.Position{
			Line:   maxRow,
			Column: maxCol,
		}
	}
	return &rg
}

func (s *Sheet) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		for _, r := range s.rows {
			row := r.Data()
			if !yield(row) {
				break
			}
		}
	}
	return it
}

func (s *Sheet) Cell(pos layout.Position) (value.Value, error) {
	cell, ok := s.cells[pos]
	if !ok {
		return formula.ErrValue, nil
	}
	return cell.Parsed, nil
}

func (s *Sheet) Copy(other *Sheet) error {
	for _, rs := range other.rows {
		s.Size.Lines++
		x := Row{
			Line:  rs.Line,
			Cells: rs.cloneCells(),
		}
		s.rows = append(other.rows, &x)
		s.Size.Columns = max(s.Size.Columns, int64(len(x.Cells)))
	}
	return nil
}

func (s *Sheet) Append(data []string) error {
	rs := Row{
		Line: int64(len(s.rows)) + 1,
	}
	s.Size.Lines++
	for i, d := range data {
		pos := layout.Position{
			Line:   rs.Line,
			Column: int64(i) + 1,
		}
		g := Cell{
			Position: pos,
			Raw:      d,
			Parsed:   formula.Text(d),
		}
		c := Cell{
			Type: TypeInlineStr,
			Cell: &g,
		}
		rs.Cells = append(rs.Cells, &c)
	}
	s.Size.Columns = max(s.Size.Columns, int64(len(data)))
	s.rows = append(s.rows, &rs)
	return nil
}

func (s *Sheet) Encode(e Encoder) error {
	return e.EncodeSheet(s)
}

type projectedView struct {
	sheet   View
	columns []int64
	mapping map[int64]int64
}

func Project(view View, sel layout.Selection) View {
	return newProjectedView(view, sel)
}

func newProjectedView(sh View, sel layout.Selection) View {
	v := projectedView{
		sheet:   sh,
		columns: sel.Indices(sh.Bounds()),
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

func (v *projectedView) Cell(pos layout.Position) (value.Value, error) {
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

func newBoundedView(sh View, rg *layout.Range) View {
	v := boundedView{
		sheet: sh,
		part:  rg.Normalize(),
	}
	return &v
}

func (v *boundedView) Name() string {
	return v.sheet.Name()
}

func (v *boundedView) Cell(pos layout.Position) (value.Value, error) {
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
				val, err := v.sheet.Cell(p)
				if err == nil {
					data[ix] = val.(value.ScalarValue)
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
