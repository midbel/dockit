package oxml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strconv"
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

type Cell struct {
	Type string
	Position

	rawValue    string
	parsedValue any
	Formula     Expr
}

func (c *Cell) Value() string {
	return c.rawValue
}

func (c *Cell) Get() any {
	return c.parsedValue
}

func (c *Cell) Refresh(ctx Context) error {
	if c.Formula == nil {
		return nil
	}
	value, err := Eval(c.Formula, ctx)
	if err == nil {
		c.parsedValue = valueToScalar(value)
		c.rawValue = valueToString(value)
	}
	return err
}

type Row struct {
	Line   int64
	Hidden bool
	Cells  []*Cell
}

func (r *Row) Data() []any {
	var ds []any
	for _, c := range r.Cells {
		ds = append(ds, c.Get())
	}
	return ds
}

func (r *Row) values() []any {
	var list []any
	for i := range r.Cells {
		list = append(list, r.Cells[i].parsedValue)
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
	Name   string
	Active bool
	Index  int
	Rows   []*Row
	Size   *Dimension

	State     SheetState
	Protected SheetProtection
}

func NewSheet(name string) *Sheet {
	s := Sheet{
		Name:   name,
		Active: false,
		State:  StateVisible,
		Size:   &Dimension{},
	}
	return &s
}

func (s *Sheet) Refresh(ctx Context) error {
	ctx = SheetContext(ctx, s)
	for _, r := range s.Rows {
		for _, c := range r.Cells {
			if err := c.Refresh(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Sheet) Bounding() Bounds {
	var bounds Bounds
	bounds.Start = Position{
		Line:   1,
		Column: 1,
	}
	bounds.End = Position{
		Line:   s.Size.Lines,
		Column: s.Size.Columns,
	}
	return bounds
}

func (s *Sheet) Iter() iter.Seq[[]any] {
	it := func(yield func([]any) bool) {
		for _, r := range s.Rows {
			row := r.Data()
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
	for _, rs := range other.Rows {
		s.Size.Lines++
		x := Row{
			Line:  rs.Line,
			Cells: rs.cloneCells(),
		}
		s.Rows = append(other.Rows, &x)
		s.Size.Columns = max(s.Size.Columns, int64(len(x.Cells)))
	}
	return nil
}

func (s *Sheet) Append(data []string) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
	}
	rs := Row{
		Line: int64(len(s.Rows)) + 1,
	}
	s.Size.Lines++
	for i, d := range data {
		pos := Position{
			Line:   rs.Line,
			Column: int64(i) + 1,
		}
		c := Cell{
			rawValue:    d,
			parsedValue: d,
			Type:        TypeInlineStr,
			Position:    pos,
		}
		rs.Cells = append(rs.Cells, &c)
	}
	s.Size.Columns = max(s.Size.Columns, int64(len(data)))
	s.Rows = append(s.Rows, &rs)
	return nil
}

func (s *Sheet) Insert(ix int64, data []any) error {
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
	fmt.Println("lock", s.Name, s.Protected)
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

func (s *Sheet) At(row, col int64) (Value, error) {
	rix := slices.IndexFunc(s.Rows, func(r *Row) bool {
		return r.Line == row
	})
	if rix < 0 {
		return nil, nil
	}
	cix := slices.IndexFunc(s.Rows[rix].Cells, func(c *Cell) bool {
		return c.Line == row && c.Column == col
	})
	if cix < 0 {
		return nil, nil
	}
	return s.Rows[rix].Cells[cix].Get(), nil
}

func (s *Sheet) resetSharedIndex(ix map[int]int) {
	for _, r := range s.Rows {
		for _, c := range r.Cells {
			if c.Type != TypeSharedStr {
				continue
			}
			x, _ := strconv.Atoi(c.rawValue)
			if n, ok := ix[x]; ok {
				c.rawValue = strconv.Itoa(n)
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
		if err := s.Refresh(SheetContext(ctx, s)); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) ActiveSheet() (*Sheet, error) {
	return nil, nil
}

func (f *File) Sheet(name string) (*Sheet, error) {
	ix := slices.IndexFunc(f.sheets, func(s *Sheet) bool {
		return s.Name == name
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
	sh.Name = newName
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
		return s.Name == name
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
	if n, ok := f.names[sheet.Name]; ok {
		f.names[sheet.Name] = n + 1
		sheet.Name = fmt.Sprintf("%s_%03d", sheet.Name, f.names[sheet.Name])
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
		if n, ok := f.names[s.Name]; ok {
			f.names[s.Name] = n + 1
			s.Name = fmt.Sprintf("%s_%03d", s.Name, f.names[s.Name])
		}
		f.sheets = append(f.sheets, s)
		s.resetSharedIndex(ix)
	}
	return nil
}
