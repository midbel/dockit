package oxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"iter"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"

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

func (c *Cell) Formula() value.Formula {
	return c.formula
}

func (c *Cell) Reload(ctx value.Context) error {
	if c.formula == nil {
		return nil
	}
	res := c.formula.Eval(ctx)
	if !value.IsScalar(res) {
       	c.parsed = value.ErrValue
	} else {
       	c.parsed = res.(value.ScalarValue)
	}
	c.raw = res.String()
	return nil
}

func typeFromValue(val value.ScalarValue) string {
	switch val.Type() {
	case value.TypeNumber:
		return TypeNumber
	case value.TypeText:
		return TypeInlineStr
	case value.TypeBool:
		return TypeBool
	case value.TypeDate:
		return TypeDate
	default:
		return TypeInlineStr
	}
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

func (r *row) Len() int {
	return len(r.Cells)
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
	name = cleanName(name)
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
			Type:     TypeError,
			Position: pos,
			raw:      "",
			parsed:   value.Empty(),
		}
	}
	return cell, nil
}

func (s *Sheet) Reload(ctx value.Context) error {
	ctx = grid.EnclosedContext(ctx, grid.SheetContext(s))
	for _, r := range s.rows {
		for _, c := range r.Cells {
			c.Reload(ctx)
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

func (s *Sheet) Copy(other grid.View) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
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
	return nil
}

func (s *Sheet) SetFormula(pos layout.Position, expr value.Formula) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.formula = expr
	c.raw = ""
	c.parsed = value.Empty()
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
	c.parsed = value.Empty()
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
		r = &row{
			Line: pos.Line,
		}
		s.rows = append(s.rows, r)
	} else {
		r = s.rows[ix]
	}
	c := &Cell{
		Type:     typeFromValue(val),
		Position: pos,
		raw:      val.String(),
		parsed:   val,
	}
	r.Cells = append(r.Cells, c)
	s.cells[pos] = c
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

	names         *grid.NameIndex
	sheets        []*Sheet
	sharedStrings []string
}

func NewFile() *File {
	file := &File{
		names: grid.NewNameIndex(),
	}
	return file
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
	ws, err := writeFile(w)
	if err != nil {
		return err
	}
	defer ws.Close()
	return ws.WriteFile(f)
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
	ctx := grid.NewContext(grid.FileContext(f))
	for _, s := range f.sheets {
		if err := s.Reload(ctx); err != nil {
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
	if err := f.RemoveSheet(oldName); err != nil {
		return err
	}
	sh.Label = cleanName(newName)
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

func (f *File) RemoveSheet(name string) error {
	if f.locked {
		return grid.ErrLock
	}
	size := len(f.sheets)
	f.sheets = slices.DeleteFunc(f.sheets, func(s *Sheet) bool {
		return s.Name() == name
	})
	if size != len(f.sheets) {
		f.names.Delete(name)
	}
	return nil
}

func (f *File) AppendSheet(sheet grid.View) error {
	if f.locked {
		return grid.ErrLock
	}
	sh := NewSheet(sheet.Name())
	sh.Label = cleanName(sheet.Name())
	sh.Label = f.names.Next(sh.Label)
	if err := sh.Copy(sheet); err != nil {
		return err
	}
	sh.Index = len(f.sheets) + 1
	sh.Id = fmt.Sprintf("rId%d", sh.Index)
	f.sheets = append(f.sheets, sh)
	return nil
}

// append sheets of given file to current fule
func (f *File) Merge(other grid.File) error {
	if f.locked {
		return grid.ErrLock
	}
	var err error
	if x, ok := other.(*File); ok {
		err = f.mergeFile(x)
	} else {
		err = f.mergeSheetsFromFile(other)
	}
	return err
}

func (f *File) mergeFile(other *File) error {
	ix := make(map[int]int)
	for i, s := range other.sharedStrings {
		ok := slices.Contains(f.sharedStrings, s)
		if ok {
			continue
		}
		ix[i] = len(f.sharedStrings)
		f.sharedStrings = append(f.sharedStrings, s)
	}
	for _, s := range other.sheets {
		if err := f.AppendSheet(s); err != nil {
			return err
		}
		s.resetSharedIndex(ix)
	}
	return nil
}

func (f *File) mergeSheetsFromFile(other grid.File) error {
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

const maxSheetNameLen = 31

func cleanName(str string) string {
	ret := grid.CleanName(str)
	if len(ret) > maxSheetNameLen {
		ret = ret[:maxSheetNameLen]
	}
	return ret
}
