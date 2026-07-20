package flat

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"os"
	"slices"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/id"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
	"github.com/midbel/log"
)

type Reader interface {
	Read() ([]string, error)
}

type File struct {
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
	}
	return &file, nil
}

func NewFile() *File {
	var file File
	return &file
}

func NewFileFromRows(rs [][]value.Value) *File {
	sh := NewSheet(defaultSheetName, rs)
	return NewFileFromSheets(sh)
}

func NewFileFromSheets(sheets ...*Sheet) *File {
	f := NewFile()
	f.sheets = append(f.sheets, sheets...)
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
	if len(f.sheets) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	return f.sheets[0], nil
}

func (f *File) Sheet(ident string) (grid.View, error) {
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

func (f *File) Rename(oldName, newName string) error {
	return nil
}

func (f *File) Copy(oldName, newName string) error {
	return nil
}

func (f *File) AppendSheet(view grid.View) error {
	sh, ok := view.(*Sheet)
	if !ok {
		sh = NewSheet(view.Name(), nil)
		if err := sh.FillWith(view); err != nil {
			return err
		}
	}
	f.sheets = append(f.sheets, sh)
	return nil
}

func (f *File) RemoveSheet(name string) error {
	return nil
}

func (f *File) CloneSheet(ident string, mode grid.CopyMode) (grid.View, error) {
	return nil, nil
}

func (f *File) Sync() error {
	ctx := grid.FileContext(f)
	for _, s := range f.sheets {
		if err := s.Sync(ctx); err != nil {
			return err
		}
	}
	return nil
}

const defaultSheetName = "sheet1"

type Sheet struct {
	Label string

	rows  []*row
	cells map[layout.Position]*Cell
	size  layout.Dimension
}

func NewSheet(name string, values [][]value.Value) *Sheet {
	s := namedSheet(name)
	for i, rs := range values {
		r := createRow(int64(i + 1))
		for j, v := range rs {
			var (
				pos  = layout.NewPosition(r.Line, int64(j+1))
				cell = valueCell(pos, v)
			)
			s.cells[cell.At()] = cell
			r.Cells = append(r.Cells, cell)
		}
		s.rows = append(s.rows, r)
	}
	return s
}

func emptySheet() *Sheet {
	return namedSheet(defaultSheetName)
}

func namedSheet(name string) *Sheet {
	return &Sheet{
		Label: name,
		cells: make(map[layout.Position]*Cell),
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
		r := createRow(int64(line))
		for col, f := range fields {
			var (
				pos  = layout.NewPosition(r.Line, int64(col)+1)
				cell = valueCell(pos, value.Text(f))
			)
			r.Cells = append(r.Cells, cell)
			sh.cells[pos] = cell
		}
		sh.rows = append(sh.rows, r)
		sh.size.Lines++
		sh.size.Columns = max(sh.size.Columns, int64(len(fields)))
	}
	return sh, nil
}

func (s *Sheet) Name() string {
	return s.Label
}

func (s *Sheet) Rename(name string) {
	s.Label = name
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

func (s *Sheet) Sync(ctx value.Context) error {
	ctx = grid.EnclosedContext(ctx, grid.SheetContext(s))
	for _, r := range s.rows {
		for _, c := range r.Cells {
			if err := c.Sync(ctx); err != nil {
				return err
			}
		}
	}
	return nil
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

func (s *Sheet) Rows() iter.Seq2[int64, []value.Value] {
	it := func(yield func(int64, []value.Value) bool) {
		for _, r := range s.rows {
			if len(r.Cells) == 0 {
				continue
			}
			res := make([]value.Value, len(r.Cells))
			for i, c := range r.Cells {
				res[i] = c.Value()
			}
			if !yield(r.Line, res) {
				return
			}
		}
	}
	return it
}

func (s *Sheet) Clone(mode grid.CopyMode) (grid.View, error) {
	return nil, nil
}

func (s *Sheet) FillWith(other grid.View) error {
	b := other.Bounds()
	for p := range b.Positions() {
		c, _ := other.Cell(p)
		s.put(c, grid.CopyAll)
	}
	return nil
}

func (s *Sheet) Cell(pos layout.Position) (grid.Cell, error) {
	cell, ok := s.cells[pos]
	if !ok {
		return grid.Empty(pos), nil
	}
	return cell, nil
}

func (s *Sheet) SetValue(pos layout.Position, val value.Value) error {
	c, ok := s.cells[pos]
	if !ok {
		c = valueCell(pos, val)
	}
	s.ClearFormula(pos)
	c.update(val)
	s.insertOrReplaceCell(c)
	return nil
}

func (s *Sheet) SetFormula(pos layout.Position, f value.Formula) error {
	cell, ok := s.cells[pos]
	if !ok {
		cell = emptyCell(pos)
	}
	cell.formula = f
	cell.raw = f.String()
	cell.parsed = nil
	cell.dirty = true

	s.insertOrReplaceCell(cell)
	return nil
}

func (s *Sheet) ClearCell(pos layout.Position) error {
	return s.ClearValue(pos)
}

func (s *Sheet) ClearValue(pos layout.Position) error {
	c, ok := s.cells[pos]
	if !ok {
		return nil
	}
	c.Clear()
	return nil
}

func (s *Sheet) ClearRange(rg *layout.Range) error {
	return nil
}

func (s *Sheet) ClearFormula(_ layout.Position) error {
	return nil
}

func (s *Sheet) RemoveRows(offset, count int64) error {
	defer s.updateDims(-count, 0)
	ix := slices.IndexFunc(s.rows, func(r *row) bool {
		return r.Line >= offset
	})
	if ix < 0 {
		return nil
	}
	for _, r := range s.rows[ix : ix+int(count)] {
		for _, c := range r.Cells {
			delete(s.cells, c.At())
		}
	}
	for _, r := range s.rows[ix+int(count):] {
		r.Line -= count
		for _, c := range r.Cells {
			c.moveLine(-count)
			s.cells[c.At()] = c
		}
	}
	if ix+int(count) >= len(s.rows) {
		count = int64(len(s.rows) - ix)
	}
	s.rows = slices.Delete(s.rows, ix, ix+int(count))
	return nil
}

func (s *Sheet) RemoveColumns(offset, count int64) error {
	defer s.updateDims(0, -count)
	for _, r := range s.rows {
		ix := slices.IndexFunc(r.Cells, func(c *Cell) bool {
			return c.Column >= offset
		})
		if ix < 0 {
			continue
		}
		for _, c := range r.Cells[ix : ix+int(count)] {
			delete(s.cells, c.At())
		}
		for _, c := range r.Cells[ix+int(count):] {
			c.moveColumn(count)
			s.cells[c.At()] = c
		}
		if ix+int(count) >= len(r.Cells) {
			count = int64(len(r.Cells) - ix)
		}
		r.Cells = slices.Delete(r.Cells, ix, ix+int(count))
	}
	return nil
}

func (s *Sheet) InsertRows(offset, count int64) error {
	rows := make([]*row, count)
	for i := range rows {
		rows[i] = createRow(offset + int64(i) + 1)
		for j := int64(1); j <= s.size.Columns; j++ {
			var (
				pos  = layout.NewPosition(rows[i].Line, j)
				cell = emptyCell(pos)
			)
			rows[i].Cells = append(rows[i].Cells, cell)
			s.cells[cell.At()] = cell
		}
	}
	defer s.updateDims(count, 0)
	if offset == 0 {
		s.rows = append(rows, s.rows...)
		for i := range s.rows[int(count):] {
			pos := s.rows[i+int(count)].shift(count)
			maps.Copy(s.cells, pos)
		}
		return nil
	}
	ix := slices.IndexFunc(s.rows, func(r *row) bool {
		return r.Line >= offset
	})
	if ix < 0 {
		s.rows = append(s.rows, rows...)
	} else {
		for i := ix + 1; i < len(s.rows); i++ {
			pos := s.rows[i].shift(count)
			maps.Copy(s.cells, pos)
		}
		s.rows = slices.Insert(s.rows, ix+1, rows...)
	}
	return nil
}

func (s *Sheet) InsertColumns(offset, count int64) error {
	defer s.updateDims(0, count)
	for i := range s.rows {
		cols := make([]*Cell, count)
		for j := int64(0); j < count; j++ {
			var (
				pos  = layout.NewPosition(s.rows[i].Line, offset+j+1)
				cell = emptyCell(pos)
			)
			cols[j] = cell
			s.cells[cell.At()] = cell
		}
		if offset == 0 {
			for _, c := range s.rows[i].Cells {
				c.Column += count
				s.cells[c.At()] = c
			}
			s.rows[i].Cells = append(cols, s.rows[i].Cells...)
			continue
		}
		ix := slices.IndexFunc(s.rows[i].Cells, func(c *Cell) bool {
			return c.Column >= offset
		})
		if ix < 0 {
			s.rows[i].Cells = append(s.rows[i].Cells, cols...)
		} else {
			s.rows[i].Cells = slices.Insert(s.rows[i].Cells, ix+1, cols...)
			for _, c := range s.rows[i].Cells[ix+1+int(count):] {
				c.Column += count
				s.cells[c.At()] = c
			}
		}
	}
	return nil
}

func (s *Sheet) insertOrReplaceCell(cell *Cell) {
	s.cells[cell.Position] = cell

	ix := slices.IndexFunc(s.rows, func(r *row) bool {
		return r.Line == cell.Line
	})
	if ix < 0 {
		r := createRow(cell.Line)
		r.Append(cell)
		s.rows = append(s.rows, r)
		slices.SortFunc(s.rows, func(r1, r2 *row) int {
			return int(r1.Line) - int(r2.Line)
		})
	} else {
		s.rows[ix].AppendOrReplace(cell)
	}

	s.updateSize(cell)
}

func (s *Sheet) updateDims(rows, cols int64) {
	s.size.Lines += rows
	s.size.Columns += cols
}

func (s *Sheet) updateSize(cell *Cell) {
	s.size.Columns = max(s.size.Columns, cell.Column)
	s.size.Lines = max(s.size.Lines, cell.Line)
}

func (s *Sheet) put(cell grid.Cell, mode grid.CopyMode) {
	var (
		pos = cell.At()
		val = cell.Value()
	)
	c := &Cell{
		id:       id.Next(),
		Position: pos,
	}
	if mode.Value() {
		c.raw = val.String()
		c.parsed = val
	}
	if f := cell.Formula(); f != nil && mode.Formula() {
		c.formula = f
	}
	s.insertOrReplaceCell(c)
}

type row struct {
	Line  int64
	Cells []*Cell
}

func createRow(lino int64) *row {
	r := &row{
		Line: lino,
	}
	return r
}

func (r *row) AppendOrReplace(cell *Cell) {
	cx := slices.IndexFunc(r.Cells, func(other *Cell) bool {
		return other.Position.Equal(cell.Position)
	})
	if cx >= 0 {
		r.Cells[cx] = cell
	} else {
		r.Append(cell)
	}
}

func (r *row) Append(cell *Cell) {
	r.Cells = append(r.Cells, cell)
	slices.SortFunc(r.Cells, func(c1, c2 *Cell) int {
		return int(c1.Column - c2.Column)
	})
}

func (r *row) Values() []value.Value {
	var ds []value.Value
	for _, c := range r.Cells {
		ds = append(ds, c.Value())
	}
	return ds
}

func (r *row) shift(count int64) map[layout.Position]*Cell {
	r.Line += count

	pos := make(map[layout.Position]*Cell)
	for _, c := range r.Cells {
		c.moveLine(count)
		pos[c.At()] = c
	}
	return pos
}

type Cell struct {
	id uint64
	layout.Position
	raw     string
	parsed  value.Value
	formula value.Formula
	dirty   bool

	link *grid.Link
}

func emptyCell(pos layout.Position) *Cell {
	return valueCell(pos, value.Empty())
}

func valueCell(pos layout.Position, val value.Value) *Cell {
	return &Cell{
		id:       id.Next(),
		Position: pos,
		raw:      val.String(),
		parsed:   val,
	}
}

func (c *Cell) DependsOn() []grid.Cell {
	return c.link.DependsOn
}

func (c *Cell) UsedBy() []grid.Cell {
	return c.link.UsedBy
}

func (c *Cell) Id() uint64 {
	return c.id
}

func (c *Cell) At() layout.Position {
	return c.Position
}

func (c *Cell) SetAt(pos layout.Position) {
	c.Position = pos
}

func (c *Cell) moveLine(count int64) {
	c.Line += count
}

func (c *Cell) moveColumn(count int64) {
	c.Column += count
}

func (c *Cell) Display() string {
	return c.raw
}

func (c *Cell) Value() value.Value {
	if c.parsed == nil {
		return value.Empty()
	}
	return c.parsed
}

func (c *Cell) Clear() {
	c.raw = ""
	c.parsed = value.Empty()
}

func (c *Cell) Formula() value.Formula {
	return c.formula
}

func (c *Cell) Dirty() bool {
	return c.dirty
}

func (c *Cell) Sync(ctx value.Context) error {
	if c.formula == nil {
		return nil
	}
	val, err := grid.Eval(c.formula, ctx)
	if err == nil {
		c.update(val)
		c.dirty = false
	}
	return err
}

func (c *Cell) update(val value.Value) error {
	if !value.IsScalar(val) {
		c.parsed = value.ErrValue
	} else {
		c.parsed = val.(value.ScalarValue)
	}
	c.raw = val.String()
	return nil
}
