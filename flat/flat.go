package flat

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"slices"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
	"github.com/midbel/log"
)

type Reader interface {
	Read() ([]string, error)
}

type Mode int

const (
	flatMode Mode = iota
	memMode
)

type File struct {
	mode   Mode
	sheets []*Sheet
}

func OpenLog(file, pattern string) (*File, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	rs, err := log.NewReader(r, pattern)
	if err != nil {
		return nil, err
	}
	return OpenReader(rs)
}

func OpenCsv(file string) (*File, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return OpenReader(csv.NewReader(r))
}

func OpenReader(r Reader) (*File, error) {
	sh, err := readSheet(r)
	if err != nil {
		return nil, err
	}
	file := File{
		sheets: []*Sheet{sh},
		mode:   flatMode,
	}
	return &file, nil
}

func NewFile() *File {
	var file File
	return &file
}

func NewFileFromRows(rs [][]value.ScalarValue) *File {
	sh := NewSheet(defaultSheetName, rs)
	return NewFileFromSheets(sh)
}

func NewFileFromSheets(sheets ...*Sheet) *File {
	f := NewFile()
	f.mode = memMode
	for _, s := range sheets {
		f.sheets = append(f.sheets, s)
	}
	return f
}

func (f *File) WriteFile(file string) error {
	w, err := os.Create(file)
	if err == nil {
		defer w.Close()
		err = f.WriteTo(w)
	}
	return err
}

func (f *File) WriteTo(w io.Writer) error {
	return nil
}

func (f *File) ActiveSheet() (grid.View, error) {
	return f.sheets[0], nil
}

func (f *File) Sheet(ident string) (grid.View, error) {
	if f.mode == flatMode {
		if ident != defaultSheetName {
			return nil, fmt.Errorf("default sheet not found")
		}
		return f.sheets[0], nil
	}
	ix := slices.IndexFunc(f.sheets, func(s *Sheet) bool {
		return s.Label == ident
	})
	if ix < 0 {
		return nil, fmt.Errorf("%s: sheet not found", ident)
	}
	return f.sheets[ix], nil
}

func (f *File) Sheets() []grid.View {
	var views []grid.View
	for i := range f.sheets {
		views = append(views, f.sheets[i])
	}
	return views
}

func (f *File) Infos() []grid.ViewInfo {
	var infos []grid.ViewInfo
	for _, s := range f.sheets {
		rg := s.Bounds()

		i := grid.ViewInfo{
			Name:      s.Name(),
			Active:    false,
			Protected: true,
			Hidden:    false,
			Size: layout.Dimension{
				Lines:   rg.Height(),
				Columns: rg.Width(),
			},
		}
		infos = append(infos, i)
	}

	return infos
}

func (f *File) Rename(_, _ string) error {
	if err := f.supported(); err != nil {
		return err
	}
	return nil
}

func (f *File) Copy(oldName, newName string) error {
	if err := f.supported(); err != nil {
		return err
	}
	return nil
}

func (f *File) AppendSheet(view grid.View) error {
	if err := f.supported(); err != nil {
		return err
	}
	return nil
}

func (f *File) RemoveSheet(name string) error {
	if err := f.supported(); err != nil {
		return err
	}
	return nil
}

func (f *File) Reload() error {
	if err := f.supported(); err != nil {
		return err
	}
	return nil
}

func (f *File) supported() error {
	if f.mode == flatMode {
		return grid.ErrSupported
	}
	return nil
}

const defaultSheetName = "sheet"

type Sheet struct {
	Label string
	mode  Mode

	rows  []*row
	cells map[layout.Position]*Cell
	size  layout.Dimension
}

func NewSheet(name string, values [][]value.ScalarValue) *Sheet {
	s := Sheet{
		Label: name,
		cells: make(map[layout.Position]*Cell),
		mode:  memMode,
	}
	for i, rs := range values {
		r := &row{
			Line: int64(i + 1),
		}
		for j, v := range rs {
			pos := layout.NewPosition(r.Line, int64(j+1))
			cell := &Cell{
				raw:      v.String(),
				parsed:   v,
				Position: pos,
			}
			s.cells[pos] = cell
			r.Cells = append(r.Cells, cell)
		}
		s.rows = append(s.rows, r)
	}
	return &s
}

func emptySheet() *Sheet {
	return &Sheet{
		Label: defaultSheetName,
		cells: make(map[layout.Position]*Cell),
		mode:  flatMode,
	}
}

func readSheet(rs Reader) (*Sheet, error) {
	sh := emptySheet()
	for line := 1; ; line++ {
		fields, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		r := &row{
			Line: int64(line),
		}
		for col, f := range fields {
			p := layout.Position{
				Line:   r.Line,
				Column: int64(col) + 1,
			}
			c := &Cell{
				Position: p,
				raw:      f,
				parsed:   value.Text(f),
			}
			r.Cells = append(r.Cells, c)
			sh.cells[p] = c
		}
		sh.rows = append(sh.rows, r)
	}
	return sh, nil
}

func (s *Sheet) Name() string {
	return s.Label
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
	if n := len(s.rows[0].Cells); n > 0 {
		start.Column = 1
		end.Column = int64(n)
	}
	return layout.NewRange(start, end)
}

func (s *Sheet) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		for _, r := range s.rows {
			if len(r.Cells) == 0 {
				continue
			}
			res := make([]value.ScalarValue, len(r.Cells))
			for i, c := range r.Cells {
				res[i] = c.Value()
			}
			if !yield(res) {
				return
			}
		}
	}
	return it
}

func (s *Sheet) Cell(pos layout.Position) (grid.Cell, error) {
	cell, ok := s.cells[pos]
	if !ok {
		cell = &Cell{
			Position: pos,
			raw:      value.ErrRef.String(),
			parsed:   value.ErrRef,
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
	if err := s.supported(); err != nil {
		return err
	}
	return nil
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
	if err := s.supported(); err != nil {
		return err
	}
	return nil
}

func (s *Sheet) ClearFormula(_ layout.Position) error {
	if err := s.supported(); err != nil {
		return err
	}
	return nil
}

func (s *Sheet) AppendRow(values []value.ScalarValue) error {
	if err := s.supported(); err != nil {
		return err
	}
	return nil
}

func (s *Sheet) InsertRow(ix int64, values []value.ScalarValue) error {
	if err := s.supported(); err != nil {
		return err
	}
	return nil
}

func (s *Sheet) DeleteRow(ix int64) error {
	if err := s.supported(); err != nil {
		return err
	}
	return nil
}

func (s *Sheet) supported() error {
	if s.mode == flatMode {
		return grid.ErrSupported
	}
	return nil
}

type row struct {
	Line  int64
	Cells []*Cell
}

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
