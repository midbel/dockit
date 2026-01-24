package oxml

import (
	"encoding/xml"
	"fmt"
	"iter"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/eval"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

const (
	FormulaNormal = "normal"
	FormulaShared = "shared"
)

const (
	TypeSharedStr = "s"
	TypeInlineStr = "inlineStr"
	TypeFormula   = "str"
	TypeDate      = "d"
	TypeError     = "e"
	TypeBool      = "b"
	TypeNumber    = "n"
)

type Cell struct {
	Type  string
	style int
	layout.Position

	raw     string
	parsed  value.ScalarValue
	formula value.Formula
	dirty   bool
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
	if c.formula == nil {
		return nil
	}
	res, err := c.formula.Eval(ctx)
	if err == nil {
		if !types.IsScalar(res) {
			c.parsed = types.ErrValue
		} else {
			c.parsed = res.(value.ScalarValue)
		}
		c.raw = res.String()
		c.dirty = false
	}
	return err
}

type row struct {
	Line   int64
	Hidden bool
	Cells  []*Cell
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

func (r *row) cloneCells() []*Cell {
	var cells []*Cell
	for i := range r.Cells {
		c := *r.Cells[i]
		cells = append(cells, &c)
	}
	return cells
}

type SheetState int8

const (
	StateVisible SheetState = 1 << iota
	StateHidden
	StateVeryHidden
)

func (s SheetState) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	attr := xml.Attr{
		Name: name,
	}
	switch s {
	case StateVisible:
		attr.Value = "visible"
	case StateHidden:
		attr.Value = "hidden"
	case StateVeryHidden:
		attr.Value = "veryHidden"
	default:
	}
	return attr, nil
}

func (s *SheetState) UnmarshalXMLAttr(attr xml.Attr) error {
	switch attr.Value {
	case "visible":
		(*s) = StateVisible
	case "hidden":
		(*s) = StateHidden
	case "veryHidden":
		(*s) = StateVeryHidden
	default:
	}
	return nil
}

type SheetProtection int16

const (
	ProtectedSheet SheetProtection = 1 << iota
	ProtectedObjects
	ProtectedScenarios
	ProtectedFormatCells
	ProtectedFormatColumns
	ProtectedFormatRows
	ProtectedDeleteColumns
	ProtectedDeleteRows
	ProtectedInsertColumns
	ProtectedInsertRows
	ProtectedSort
	ProtectedAll
)

func (p SheetProtection) Locked() bool {
	return p&ProtectedSheet > 0
}

func (p SheetProtection) RowsLocked() bool {
	if p.Locked() {
		return true
	}
	return p&ProtectedDeleteRows > 0 || p&ProtectedInsertRows > 0
}

func (p SheetProtection) ColumnsLocked() bool {
	if p.Locked() {
		return true
	}
	return p&ProtectedDeleteColumns > 0 || p&ProtectedInsertColumns > 0
}

type Sheet struct {
	Id     string
	Label  string
	Active bool
	Index  int
	Size   layout.Dimension

	Charts []*grid.Chart

	rows  []*row
	cells map[layout.Position]*Cell

	State     SheetState
	Protected SheetProtection
}

func NewSheet(name string) *Sheet {
	name = cleanSheetName(name)
	s := Sheet{
		Label:  name,
		Active: false,
		State:  StateVisible,
		cells:  make(map[layout.Position]*Cell),
	}
	return &s
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
	cell, ok := s.cells[pos]
	if !ok {
		cell = &Cell{
			Type:     TypeInlineStr,
			Position: pos,
			raw:      "",
			parsed:   types.Blank{},
		}
	}
	return cell, nil
}

func (s *Sheet) Reload(ctx value.Context) error {
	ctx = eval.SheetContext(ctx, s)
	for _, r := range s.rows {
		for _, c := range r.Cells {
			if err := c.Reload(ctx); err != nil {
				return err
			}
		}
	}
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

func (s *Sheet) Copy(other *Sheet) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return grid.ErrLock
	}
	for _, rs := range other.rows {
		s.Size.Lines++
		x := row{
			Line:  rs.Line,
			Cells: rs.cloneCells(),
		}
		s.rows = append(other.rows, &x)
		s.Size.Columns = max(s.Size.Columns, int64(len(x.Cells)))
	}
	return nil
}

func (s *Sheet) Encode(e grid.Encoder) error {
	return e.EncodeSheet(s)
}

func (s *Sheet) Lock() {
	s.Protected = ProtectedAll - 1
}

func (s *Sheet) Unlock() {
	s.Protected = 0
}

func (s *Sheet) IsLock() bool {
	return s.Protected != 0
}

func (s *Sheet) SetValue(pos layout.Position, val value.ScalarValue) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	if err := s.ClearFormula(pos); err != nil {
		return err
	}
	c.raw = val.String()
	c.parsed = val
	c.dirty = false
	return nil
}

func (s *Sheet) SetFormula(pos layout.Position, expr value.Formula) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.formula = expr
	c.raw = ""
	c.parsed = types.Blank{}
	c.dirty = true
	return nil
}

func (s *Sheet) ClearCell(pos layout.Position) error {
	err := s.ClearValue(pos)
	if err != nil {
		return err
	}
	return s.ClearFormula(pos)
}

func (s *Sheet) ClearValue(pos layout.Position) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.raw = ""
	c.parsed = types.Blank{}
	c.dirty = false
	return nil
}

func (s *Sheet) ClearRange(rg *layout.Range) error {
	return nil
}

func (s *Sheet) ClearFormula(pos layout.Position) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.formula = nil
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

func (s *Sheet) resetSharedIndex(ix map[int]int) {
	for _, r := range s.rows {
		for _, c := range r.Cells {
			if c.Type != TypeSharedStr {
				continue
			}
			x, _ := strconv.Atoi(c.raw)
			if n, ok := ix[x]; ok {
				c.raw = strconv.Itoa(n)
			}
		}
	}
}

type File struct {
	locked   bool
	date1904 bool

	names         map[string]int
	sheets        []*Sheet
	sharedStrings []string
}

func NewFile() *File {
	var file File
	file.names = make(map[string]int)
	return &file
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
	w, err := writeFile(file)
	if err != nil {
		return err
	}
	defer w.Close()
	return w.WriteFile(f)
}

func (f *File) Infos() []grid.ViewInfo {
	var infos []grid.ViewInfo
	for _, s := range f.sheets {
		i := grid.ViewInfo{
			Name:      s.Name(),
			Active:    s.Active,
			Protected: s.IsLock(),
			Hidden:    s.State != StateVisible,
			Size:      s.Size,
		}

		infos = append(infos, i)
	}
	return infos
}

func (f *File) Reload() error {
	ctx := eval.FileContext(env.Empty(), f)
	for _, s := range f.sheets {
		if err := s.Reload(eval.SheetContext(ctx, s)); err != nil {
			return err
		}
	}
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

func (f *File) Lock() {
	for i := range f.sheets {
		f.sheets[i].Lock()
	}
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

func (f *File) UnlockSheet(name string) error {
	sh, err := f.sheetByName(name)
	if err == nil {
		sh.Unlock()
	}
	return err
}

// rename a sheet
func (f *File) Rename(oldName, newName string) error {
	if f.locked {
		return grid.ErrLock
	}
	sh, err := f.sheetByName(oldName)
	if err != nil {
		return err
	}
	if err := f.Remove(oldName); err != nil {
		return err
	}
	sh.Label = cleanSheetName(newName)
	return f.AppendSheet(sh)
}

// copy a sheet
func (f *File) Copy(oldName, newName string) error {
	if f.locked {
		return grid.ErrLock
	}
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

func (f *File) Remove(name string) error {
	if f.locked {
		return grid.ErrLock
	}
	size := len(f.sheets)
	f.sheets = slices.DeleteFunc(f.sheets, func(s *Sheet) bool {
		return s.Name() == name
	})
	if n, ok := f.names[name]; ok && n == 1 && len(f.sheets) < size {
		delete(f.names, name)
	}
	return nil
}

func (f *File) AppendSheet(sheet *Sheet) error {
	if f.locked {
		return grid.ErrLock
	}
	sheet.Label = cleanSheetName(sheet.Label)
	if n, ok := f.names[sheet.Label]; ok {
		f.names[sheet.Label] = n + 1
		sheet.Label = fmt.Sprintf("%s_%03d", sheet.Label, f.names[sheet.Label])
	}
	sheet.Index = len(f.sheets) + 1
	sheet.Id = fmt.Sprintf("rId%d", sheet.Index)
	f.sheets = append(f.sheets, sheet)
	return nil
}

// append sheets of given file to current fule
func (f *File) Merge(other *File) error {
	if f.locked {
		return grid.ErrLock
	}
	ix := make(map[int]int)
	for i, s := range other.sharedStrings {
		ok := slices.Contains(f.sharedStrings, s)
		if ok {
			continue
		}
		ix[i] = len(f.sharedStrings)
		f.sharedStrings = append(f.sharedStrings, s)
	}
	for i, s := range other.sheets {
		s.Index = len(f.sheets) + i + 1
		s.Id = fmt.Sprintf("rId%d", s.Index)

		s.Label = cleanSheetName(s.Label)
		if n, ok := f.names[s.Label]; ok {
			f.names[s.Label] = n + 1
			s.Label = fmt.Sprintf("%s_%03d", s.Label, f.names[s.Label])
		}
		f.sheets = append(f.sheets, s)
		s.resetSharedIndex(ix)
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
		return nil, fmt.Errorf("missing active sheet")
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

func cleanSheetName(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			return r
		}
		return -1
	}, str)
}
