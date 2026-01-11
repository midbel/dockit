package oxml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"iter"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/midbel/dockit/formula"
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

var (
	ErrFile        = errors.New("invalid spreadsheet")
	ErrLock        = errors.New("spreadsheet locked")
	ErrFound       = errors.New("not found")
	ErrImplemented = errors.New("not implemented")
)

type Cell struct {
	Type  string
	style int
	layout.Position

	raw     string
	parsed  value.ScalarValue
	formula formula.Expr
	dirty   bool
}

func (c *Cell) Display() string {
	return c.raw
}

func (c *Cell) Value() value.ScalarValue {
	return c.parsed
}

func (c *Cell) Reload(ctx formula.Context) error {
	if c.formula == nil {
		return nil
	}
	res, err := formula.Eval(c.formula, ctx)
	if err == nil {
		if !formula.IsScalar(res) {
			c.parsed = formula.ErrValue
		} else {
			c.parsed = res.(value.ScalarValue)
		}
		c.raw = res.String()
		c.dirty = false
	}
	return err
}

type Row struct {
	Line   int64
	Hidden bool
	Cells  []*Cell
}

func (r *Row) Values() []value.ScalarValue {
	var ds []value.ScalarValue
	for _, c := range r.Cells {
		ds = append(ds, c.Value())
	}
	return ds
}

func (r *Row) Sparse() bool {
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

func (r *Row) cloneCells() []*Cell {
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

type View interface {
	Name() string
	Bounds() *layout.Range
	Cell(layout.Position) (*Cell, error)
	Cells() iter.Seq[*Cell]
	Rows() iter.Seq[[]value.ScalarValue]
	Encode(Encoder) error
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

func (v *projectedView) Cell(pos layout.Position) (*Cell, error) {
	if pos.Column < 0 || pos.Column > int64(len(v.columns)) {
		return nil, nil
	}
	mod := layout.Position{
		Column: v.columns[pos.Column],
		Line:   pos.Line,
	}
	return v.sheet.Cell(mod)
}

func (v *projectedView) Cells() iter.Seq[*Cell] {
	it := func(yield func(*Cell) bool) {
		for c := range v.sheet.Cells() {
			col, ok := v.mapping[c.Position.Column]
			if !ok {
				continue
			}

			c.Position.Column = col
			if !yield(c) {
				return
			}
		}
	}
	return it
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

func (v *boundedView) Cell(pos layout.Position) (*Cell, error) {
	if !v.part.Contains(pos) {
		return nil, fmt.Errorf("position outside view range")
	}
	return v.sheet.Cell(pos)
}

func (v *boundedView) Bounds() *layout.Range {
	return v.part
}

func (v *boundedView) Cells() iter.Seq[*Cell] {
	it := func(yield func(*Cell) bool) {
		for c := range v.sheet.Cells() {
			if !v.part.Contains(c.Position) {
				continue
			}
			if ok := yield(c); !ok {
				return
			}
		}
	}
	return it
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

type Sheet struct {
	Id     string
	Label  string
	Active bool
	Index  int
	Size   layout.Dimension

	rows  []*Row
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

func (s *Sheet) View(rg *layout.Range) View {
	bd := s.Bounds()
	rg.Starts = rg.Starts.Update(bd.Starts)
	rg.Ends = rg.Ends.Update(bd.Ends)
	return newBoundedView(s, rg)
}

func (s *Sheet) Sub(start, end layout.Position) View {
	return s.View(layout.NewRange(start, end))
}

func (s *Sheet) Cell(pos layout.Position) (*Cell, error) {
	cell, ok := s.cells[pos]
	if !ok {
		cell = &Cell{
			Type:     TypeInlineStr,
			Position: pos,
			raw:      "",
			parsed:   formula.Blank{},
		}
	}
	return cell, nil
}

func (s *Sheet) Reload(ctx formula.Context) error {
	ctx = SheetContext(ctx, s)
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
	for c := range s.Cells() {
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

func (s *Sheet) Cells() iter.Seq[*Cell] {
	return maps.Values(s.cells)
}

func (s *Sheet) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		for _, r := range s.rows {
			row := r.Values()
			if !yield(row) {
				break
			}
		}
	}
	return it
}

func (s *Sheet) Copy(other *Sheet) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
	}
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
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
	}
	rs := Row{
		Line: int64(len(s.rows)) + 1,
	}
	s.Size.Lines++
	for i, d := range data {
		pos := layout.Position{
			Line:   rs.Line,
			Column: int64(i) + 1,
		}
		c := Cell{
			raw:      d,
			parsed:   formula.Text(d),
			Type:     TypeInlineStr,
			Position: pos,
		}
		rs.Cells = append(rs.Cells, &c)
	}
	s.Size.Columns = max(s.Size.Columns, int64(len(data)))
	s.rows = append(s.rows, &rs)
	return nil
}

func (s *Sheet) Insert(pos layout.Position, data []any) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
	}
	return nil
}

func (s *Sheet) Encode(e Encoder) error {
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

func (s *Sheet) Status() string {
	if s.State <= StateVisible {
		return "visible"
	}
	return "hidden"
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

func (f *File) Reload() error {
	ctx := FileContext(f)
	for _, s := range f.sheets {
		if err := s.Reload(SheetContext(ctx, s)); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) ActiveSheet() (*Sheet, error) {
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

func (f *File) Sheet(name string) (*Sheet, error) {
	ix := slices.IndexFunc(f.sheets, func(s *Sheet) bool {
		return s.Name() == name
	})
	if ix < 0 {
		return nil, fmt.Errorf("sheet %s %w", name, ErrFound)
	}
	return f.sheets[ix], nil
}

func (f *File) Sheets() []*Sheet {
	return slices.Clone(f.sheets)
}

func (f *File) Lock() {
	for i := range f.sheets {
		f.sheets[i].Lock()
	}
}

func (f *File) LockSheet(name string) error {
	sh, err := f.Sheet(name)
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
	sh, err := f.Sheet(name)
	if err == nil {
		sh.Unlock()
	}
	return err
}

// rename a sheet
func (f *File) Rename(oldName, newName string) error {
	if f.locked {
		return ErrLock
	}
	sh, err := f.Sheet(oldName)
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
		return ErrLock
	}
	source, err := f.Sheet(oldName)
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
		return ErrLock
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
		return ErrLock
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
		return ErrLock
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

func (f *File) setSheetName(sheet *Sheet) error {
	return nil
}

func cleanSheetName(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			return r
		}
		return -1
	}, str)
}
