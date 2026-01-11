package oxml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/midbel/dockit/formula"
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

var (
	ErrFile        = errors.New("invalid spreadsheet")
	ErrLock        = errors.New("spreadsheet locked")
	ErrFound       = errors.New("not found")
	ErrImplemented = errors.New("not implemented")
)

type Cell struct {
	Type string
	*grid.Cell
}

func (c *Cell) Value() string {
	return c.Raw
}

func (c *Cell) Get() any {
	return c.Parsed.Scalar()
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

	*grid.Sheet

	State     SheetState
	Protected SheetProtection
}

func NewSheet(name string) *Sheet {
	name = cleanSheetName(name)
	s := Sheet{
		Label:  name,
		Active: false,
		State:  StateVisible,
		Sheet:  grid.Empty(name),
	}
	return &s
}

func (s *Sheet) Reload(ctx formula.Context) error {
	ctx = SheetContext(ctx, s)
	return s.Sheet.Reload(ctx)
}

func (s *Sheet) Copy(other *Sheet) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
	}
	return s.Sheet.Copy(other)
}

func (s *Sheet) Append(data []string) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
	}
	return s.Sheet.Append(data)
}

func (s *Sheet) Insert(pos layout.Position, data []any) error {
	if s.Protected.RowsLocked() || s.Protected.ColumnsLocked() {
		return ErrLock
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
			x, _ := strconv.Atoi(c.Raw)
			if n, ok := ix[x]; ok {
				c.Raw = strconv.Itoa(n)
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
