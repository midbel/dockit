package flat

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Reader interface {
	Read() ([]string, error)
}

const defaultSheetName = "sheet"

type Cell struct {
	layout.Position
	raw    string
	parsed value.ScalarValue
}

func (c *Cell) At() layout.Position {
	return c.Position
}

func (c *Cell) Display() string {
	return c.raw
}

func (c *Cell) Value() value.ScalarValue {
	return c.parsed
}

func (c *Cell) Reload(ctx value.Context) error {
	return grid.ErrSupported
}

type row struct {
	Line  int64
	cells []*Cell
}

type Sheet struct {
	rows  []*row
	cells map[layout.Position]*Cell
	size  layout.Dimension
}

func emptySheet() *Sheet {
	return &Sheet{
		cells: make(map[layout.Position]*Cell),
	}
}

func (s *Sheet) Name() string {
	return defaultSheetName
}

func (s *Sheet) Reload(_ value.Context) error {
	return grid.ErrSupported
}

func (s *Sheet) View(rg *layout.Range) grid.View {
	bd := s.Bounds()
	rg.Starts = rg.Starts.Update(bd.Starts)
	rg.Ends = rg.Ends.Update(bd.Ends)
	return grid.NewBoundedView(s, rg)
}

func (s *Sheet) Sub(start, end layout.Position) grid.View {
	return s.View(layout.NewRange(start, end))
}

func (s *Sheet) Bounds() *layout.Range {
	var (
		start layout.Position
		end   layout.Position
	)
	if len(s.rows) == 0 {
		return layout.NewRange(start, end)
	}
	start.Line = 1
	end.Line = int64(len(s.rows))
	if n := len(s.rows[0].cells); n > 0 {
		start.Column = 1
		end.Column = int64(n)
	}
	return layout.NewRange(start, end)
}

func (s *Sheet) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		for _, r := range s.rows {
			if len(r.cells) == 0 {
				continue
			}
			res := make([]value.ScalarValue, len(r.cells))
			for i, c := range r.cells {
				res[i] = c.Value()
			}
			if !yield(res) {
				return
			}
		}
	}
	return it
}

func (s *Sheet) Encode(encoder grid.Encoder) error {
	return encoder.EncodeSheet(s)
}

func (s *Sheet) Cell(pos layout.Position) (grid.Cell, error) {
	cell, ok := s.cells[pos]
	if !ok {
		cell = &Cell{
			Position: pos,
			raw:      "",
			parsed:   value.Empty(),
		}
	}
	return cell, nil
}

func (s *Sheet) SetValue(pos layout.Position, val value.ScalarValue) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.raw = val.String()
	c.parsed = val
	return nil
}

func (s *Sheet) SetFormula(_ layout.Position, _ value.Formula) error {
	return grid.ErrSupported
}

func (s *Sheet) ClearCell(pos layout.Position) error {
	return s.ClearValue(pos)
}

func (s *Sheet) ClearValue(pos layout.Position) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.raw = ""
	c.parsed = value.Empty()
	return nil
}

func (s *Sheet) ClearRange(rg *layout.Range) error {
	return nil
}

func (s *Sheet) ClearFormula(_ layout.Position) error {
	return grid.ErrSupported
}

func (s *Sheet) AppendRow(values []value.ScalarValue) error {
	return nil
}

func (s *Sheet) InsertRow(ix int64, values []value.ScalarValue) error {
	return nil
}

func (s *Sheet) DeleteRow(ix int64) error {
	return nil
}

type File struct {
	sheet *Sheet
}

func NewFile() *File {
	file := File{
		sheet: emptySheet(),
	}
	return &file
}

func Open(r Reader) (*File, error) {
	return nil, nil
}

func (f *File) WriteFile(file string) error {
	return nil
}

func (f *File) ActiveSheet() (grid.View, error) {
	return f.sheet, nil
}

func (f *File) Sheet(ident string) (grid.View, error) {
	if ident != defaultSheetName {
		return nil, fmt.Errorf("sheet not found")
	}
	return f.sheet, nil
}

func (f *File) Sheets() []grid.View {
	return []grid.View{f.sheet}
}

func (f *File) Infos() []grid.ViewInfo {
	rg := f.sheet.Bounds()

	i := grid.ViewInfo{
		Name:      f.sheet.Name(),
		Active:    true,
		Protected: false,
		Hidden:    false,
		Size: layout.Dimension{
			Lines:   rg.Height(),
			Columns: rg.Width(),
		},
	}
	return []grid.ViewInfo{i}
}

func (*File) Rename(_, _ string) error {
	return grid.ErrSupported
}

func (*File) Copy(_, _ string) error {
	return grid.ErrSupported
}

func (*File) Remove(_ string) error {
	return grid.ErrSupported
}

func (*File) Reload() error {
	return nil
}
