package oxml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
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
	rawValue    string
	cachedValue string
	parsedValue any
	Type        string
	Position
}

func (c *Cell) Value() string {
	return c.rawValue
}

func (c *Cell) Get() any {
	return c.parsedValue
}

func (c *Cell) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	el := struct {
		XMLName     xml.Name `xml:"c"`
		Addr        string   `xml:"r,attr"`
		Type        string   `xml:"t,attr"`
		RawValue    any      `xml:"v,omitempty"`
		InlineValue any      `xml:"is>t,omitempty"`
		Formula     any      `xml:"f"`
	}{
		Addr: c.Addr(),
		Type: c.Type,
	}
	if c.Type == TypeInlineStr {
		el.InlineValue = c.rawValue
	} else if c.Type == TypeFormula {
		if c.parsedValue != "" {
			el.Formula = c.parsedValue
		}
		el.RawValue = c.rawValue
	} else {
		el.RawValue = c.rawValue
	}
	return encoder.EncodeElement(&el, start)
}

func (c *Cell) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	el := struct {
		XMLName xml.Name `xml:"c"`
		Addr    string   `xml:"r,attr"`
		Type    string   `xml:"t,attr"`
		Value   string   `xml:"v"`
		Inline  string   `xml:"is>t"`
		Formula string   `xml:"f"`
	}{}
	if err := decoder.DecodeElement(&el, &start); err != nil {
		return err
	}
	c.Position = parsePosition(el.Addr)
	c.rawValue = el.Value
	c.Type = el.Type

	switch el.Type {
	case TypeInlineStr:
		c.parsedValue = el.Inline
		c.rawValue = el.Inline
	case TypeSharedStr:
	case TypeFormula:
		c.parsedValue = el.Formula
	case TypeBool:
		b, _ := strconv.ParseBool(el.Value)
		c.parsedValue = b
	case TypeNumber, "":
		n, _ := strconv.ParseFloat(el.Value, 64)
		c.parsedValue = n
	case TypeDate:
		t, _ := ParseDate(el.Value)
		c.parsedValue = t
	case TypeError:
	default:
	}
	return nil
}

type Row struct {
	Line  int64   `xml:"r,attr"`
	Cells []*Cell `xml:"c"`
}

func (r *Row) Data() []string {
	var ds []string
	for _, c := range r.Cells {
		ds = append(ds, c.Value())
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

type Dimension struct {
	Lines   int64
	Columns int64
}

func (d *Dimension) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	el := struct {
		Ref string `xml:"ref,attr"`
	}{}
	if err := decoder.DecodeElement(&el, &start); err != nil {
		return err
	}
	startIx, endIx, ok := strings.Cut(el.Ref, ":")
	if ok {
		var (
			start = parsePosition(startIx)
			end   = parsePosition(endIx)
		)
		d.Lines = (end.Line - start.Line) + 1
		d.Columns = (end.Column - start.Column) + 1
	}
	return nil
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

func (p SheetProtection) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if p == 0 {
		return nil
	}
	el := struct {
		XMLName     xml.Name `xml:"sheetProtection"`
		Sheet       any      `xml:"sheet,attr,omitempty"`
		Objects     any      `xml:"objects,attr,omitempty"`
		Scenarios   any      `xml:"scenarios,attr,omitempty"`
		FormatCells any      `xml:"formatCells,attr,omitempty"`
		DeleteCols  any      `xml:"deleteColumns,attr,omitempty"`
		InsertCols  any      `xml:"insertColumns,attr,omitempty"`
		FormatCols  any      `xml:"formatColumns,attr,omitempty"`
		DeleteRows  any      `xml:"deleteColumns,attr,omitempty"`
		InsertRows  any      `xml:"insertColumns,attr,omitempty"`
		FormatRows  any      `xml:"formatRows,attr,omitempty"`
		Sort        any      `xml:"sort,attr,omitempty"`
	}{}
	if p&ProtectedSheet != 0 {
		el.Sheet = 1
	}
	if p&ProtectedObjects != 0 {
		el.Objects = 1
	}
	if p&ProtectedScenarios != 0 {
		el.Scenarios = 1
	}
	if p&ProtectedFormatCells != 0 {
		el.FormatCells = 1
	}
	if p&ProtectedFormatColumns != 0 {
		el.FormatCols = 1
	}
	if p&ProtectedDeleteColumns != 0 {
		el.DeleteCols = 1
	}
	if p&ProtectedInsertColumns != 0 {
		el.InsertCols = 1
	}
	if p&ProtectedDeleteRows != 0 {
		el.DeleteRows = 1
	}
	if p&ProtectedInsertRows != 0 {
		el.InsertRows = 1
	}
	if p&ProtectedSort != 0 {
		el.Sort = 1
	}
	return encoder.EncodeElement(&el, start)
}

func (p *SheetProtection) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	el := struct {
		XMLName     xml.Name `xml:"sheetProtection"`
		Sheet       any      `xml:"sheet,attr,omitempty"`
		Objects     any      `xml:"objects,attr,omitempty"`
		Scenarios   any      `xml:"scenarios,attr,omitempty"`
		FormatCells any      `xml:"formatCells,attr,omitempty"`
		DeleteCols  any      `xml:"deleteColumns,attr,omitempty"`
		InsertCols  any      `xml:"insertColumns,attr,omitempty"`
		FormatCols  any      `xml:"formatColumns,attr,omitempty"`
		DeleteRows  any      `xml:"deleteColumns,attr,omitempty"`
		InsertRows  any      `xml:"insertColumns,attr,omitempty"`
		FormatRows  any      `xml:"formatRows,attr,omitempty"`
		Sort        any      `xml:"sort,attr,omitempty"`
	}{}
	if err := decoder.DecodeElement(&el, &start); err != nil {
		return err
	}
	if el.Sheet == 1 {
		(*p) |= ProtectedSheet
	}
	if el.Objects == 1 {
		(*p) |= ProtectedObjects
	}
	if el.Scenarios == 1 {
		(*p) |= ProtectedScenarios
	}
	if el.FormatCells == 1 {
		(*p) |= ProtectedFormatCells
	}
	if el.DeleteCols == 1 {
		(*p) |= ProtectedDeleteColumns
	}
	if el.InsertCols == 1 {
		(*p) |= ProtectedInsertColumns
	}
	if el.FormatCols == 1 {
		(*p) |= ProtectedFormatColumns
	}
	if el.DeleteRows == 1 {
		(*p) |= ProtectedDeleteRows
	}
	if el.InsertRows == 1 {
		(*p) |= ProtectedInsertRows
	}
	if el.FormatRows == 1 {
		(*p) |= ProtectedFormatRows
	}
	if el.Sort == 1 {
		(*p) |= ProtectedSort
	}
	return nil
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

func (s *Sheet) Extract(sel *Select) *Sheet {
	copy := NewSheet(fmt.Sprintf("%s.copy", s.Name))
	for vs := range s.Select(sel) {
		copy.Append(vs)
	}
	return copy
}

func (s *Sheet) DistinctValues(sel *Select) iter.Seq[[]string] {
	it := func(yield func([]string) bool) {
		seen := make(map[string]struct{})
		_ = seen
	}
	return it
}

func (s *Sheet) Select(sel *Select) iter.Seq[[]string] {
	it := func(yield func([]string) bool) {
		for _, rs := range s.Rows {
			vs := sel.Select(rs)
			if !yield(vs) {
				break
			}
		}
	}
	return it
}

func (s *Sheet) Iter() iter.Seq[[]string] {
	it := func(yield func([]string) bool) {
		for _, r := range s.Rows {
			row := r.Data()
			if !yield(row) {
				break
			}
		}
	}
	return it
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

type File struct {
	locked   bool
	date1904 bool

	names        map[string]int
	sheets       []*Sheet
	sharedString []string
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
	return rs.ReadFile()
}

func (f *File) WriteFile(file string) error {
	w, err := writeFile(file)
	if err != nil {
		return err
	}
	defer w.Close()
	return w.WriteFile(f)
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
	return f.sheets
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
	return nil
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
	for _, rs := range source.Rows {
		x := Row{
			Line:  rs.Line,
			Cells: rs.cloneCells(),
		}
		target.Rows = append(target.Rows, &x)
	}
	return f.AppendSheet(target)
}

func (f *File) Remove(name string) error {
	if f.locked {
		return ErrLock
	}
	slices.DeleteFunc(f.sheets, func(s *Sheet) bool {
		return s.Name == name
	})
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
	for i, s := range other.sheets {
		s.Index = len(f.sheets) + i + 1
		s.Id = fmt.Sprintf("rId%d", s.Index)
		if n, ok := f.names[s.Name]; ok {
			f.names[s.Name] = n + 1
			s.Name = fmt.Sprintf("%s_%03d", s.Name, f.names[s.Name])
		}
		f.sheets = append(f.sheets, s)
	}
	return nil
}
