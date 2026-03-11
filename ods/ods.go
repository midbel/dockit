package ods

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"math"
	"os"
	"slices"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Cell struct {
	layout.Position

	raw     string
	parsed  value.ScalarValue
	formula value.Formula
}

func (c *Cell) At() layout.Position {
	return c.Position
}

func (c *Cell) Display() string {
	return c.raw
}

func (c *Cell) Value() value.ScalarValue {
	if c.parsed == nil {
		return value.Empty()
	}
	return c.parsed
}

func (c *Cell) Reload(ctx value.Context) error {
	return nil
}

type row struct {
	Line int64
	Cells []*Cell
}

func (r *row) Values() []value.ScalarValue {
	var ds []value.ScalarValue
	for _, c := range r.Cells {
		ds = append(ds, c.Value())
	}
	return ds
}

func (r *row) Sparse() bool {
	for i := range r.Cells {
		if i == 0 {
			continue
		}
		if r.Cells[i].Column-r.Cells[i-1].Column > 1 {
			return true
		}
	}
	return false
}

func (r *row) Clone() *row {
	var other row
	for i := range r.Cells {
		c := *r.Cells[i]
		other.Cells = append(other.Cells, &c)
	}
	return &other
}

func (r *row) Len() int {
	return len(r.Cells)
}

type Sheet struct {
	Label   string
	Active  bool
	Visible bool
	Locked  bool
	Size    layout.Dimension

	rows  []*row
	cells map[layout.Position]*Cell
}

func NewSheet(name string) *Sheet {
	sh := Sheet{
		Label:  name,
		Active: false,
		cells:  make(map[layout.Position]*Cell),
	}
	return &sh
}

func (s *Sheet) Name() string {
	return s.Label
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

func (s *Sheet) Cell(pos layout.Position) (grid.Cell, error) {
	cell, _ := s.cells[pos]
	return cell, nil
}

func (s *Sheet) Reload(ctx value.Context) error {
	return nil
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
			row := r.Values()
			if len(r.Cells) == 0 {
				continue
			}
			if !yield(row) {
				break
			}
		}
	}
	return it
}

func (s *Sheet) Copy(other grid.View) error {
	if s.Locked {
		return grid.ErrLock
	}
	b := other.Bounds()
	for p := range b.Positions() {
		c, _ := other.Cell(p)
		s.put(c)
	}
	return nil
}

func (s *Sheet) Lock() {
	s.Locked = true
}

func (s *Sheet) Unlock() {
	s.Locked = false
}

func (s *Sheet) IsLock() bool {
	return s.Locked
}

func (s *Sheet) SetValue(pos layout.Position, val value.ScalarValue) error {
	return nil
}

func (s *Sheet) SetFormula(pos layout.Position, expr value.Formula) error {
	return nil
}

func (s *Sheet) ClearCell(pos layout.Position) error {
	return nil
}

func (s *Sheet) ClearValue(pos layout.Position) error {
	return nil
}

func (s *Sheet) ClearRange(rg *layout.Range) error {
	return nil
}

func (s *Sheet) ClearFormula(pos layout.Position) error {
	return nil
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

func (s *Sheet) put(cell grid.Cell) {
	var (
		pos = cell.At()
		val = cell.Value()
	)
	ix := slices.IndexFunc(s.rows, func(r *row) bool {
		return r.Line == pos.Line
	})
	var r *row
	if ix < 0 {
		r = &row{}
		s.rows = append(s.rows, r)
	} else {
		r = s.rows[ix]
	}
	c := &Cell{
		Position: pos,
		raw:      val.String(),
		parsed:   val,
	}
	r.Cells = append(r.Cells, c)
	s.cells[pos] = c
}

type File struct {
	names  *grid.NameIndex
	sheets []*Sheet
}

func NewFile() *File {
	return &File{
		names: grid.NewNameIndex(),
	}
}

func Open(file string) (*File, error) {
	rs, err := readFile(file)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	book, err := rs.ReadFile()
	if err != nil {
		return nil, err
	}
	return book, nil
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

func (f *File) Infos() []grid.ViewInfo {
	var infos []grid.ViewInfo
	for _, s := range f.sheets {
		i := grid.ViewInfo{
			Name:      s.Name(),
			Active:    false,
			Protected: s.Locked,
			Hidden:    !s.Visible,
			Size:      s.Size,
		}

		infos = append(infos, i)
	}
	return infos
}

func (f *File) Reload() error {
	return nil
}

func (f *File) ActiveSheet() (grid.View, error) {
	return f.activeSheet()
}

func (f *File) Sheet(name string) (grid.View, error) {
	return f.sheetByName(name)
}

func (f *File) Sheets() []grid.View {
	n := len(f.sheets)
	if n == 0 {
		return nil
	}
	views := make([]grid.View, 0, n)
	for i := range f.sheets {
		views = append(views, f.sheets[i])
	}
	return views
}

func (f *File) LockSheet(name string) error {
	sh, err := f.sheetByName(name)
	if err == nil {
		sh.Lock()
	}
	return err
}

func (f *File) Unlock() {
	for i := range f.sheets {
		f.sheets[i].Unlock()
	}
}

// rename a sheet
func (f *File) Rename(oldName, newName string) error {
	sh, err := f.sheetByName(oldName)
	if err != nil {
		return err
	}
	if err := f.RemoveSheet(oldName); err != nil {
		return err
	}
	sh.Label = grid.CleanName(newName)
	return f.AppendSheet(sh)
}

// copy a sheet
func (f *File) Copy(oldName, newName string) error {
	source, err := f.sheetByName(oldName)
	if err != nil {
		return err
	}
	if newName == "" {
		newName = oldName
	}
	target := NewSheet(newName)
	target.Copy(source)
	return f.AppendSheet(target)
}

func (f *File) RemoveSheet(name string) error {
	size := len(f.sheets)
	f.sheets = slices.DeleteFunc(f.sheets, func(s *Sheet) bool {
		return s.Name() == name && !s.Locked
	})
	if size != len(f.sheets) {
		f.names.Delete(name)
	}
	return nil
}

func (f *File) AppendSheet(sheet grid.View) error {
	sh := NewSheet(sheet.Name())

	sh.Label = grid.CleanName(sheet.Name())
	sh.Label = f.names.Next(sh.Label)
	if err := sh.Copy(sheet); err != nil {
		return err
	}
	f.sheets = append(f.sheets, sh)
	return nil
}

// append sheets of given file to current fule
func (f *File) Merge(other grid.File) error {
	for _, s := range other.Sheets() {
		if err := f.AppendSheet(s); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) activeSheet() (*Sheet, error) {
	if len(f.sheets) == 1 {
		return f.sheets[0], nil
	}
	ix := slices.IndexFunc(f.sheets, func(s *Sheet) bool {
		return s.Active == true
	})
	if ix < 0 {
		return f.sheets[0], nil
	}
	return f.sheets[ix], nil
}

func (f *File) sheetByName(name string) (*Sheet, error) {
	ix := slices.IndexFunc(f.sheets, func(s *Sheet) bool {
		return s.Name() == name
	})
	if ix < 0 {
		return nil, fmt.Errorf("sheet %s %w", name, grid.ErrFound)
	}
	return f.sheets[ix], nil
}
