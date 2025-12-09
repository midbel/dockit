package oxml

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"iter"
	"path"
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

const (
	typeSheetUrl = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"
	typeDocUrl   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	typeMainUrl  = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
)

const (
	mimeRels         = "application/vnd.openxmlformats-package.relationships+xml"
	mimeXml          = "application/xml"
	mimeWorkbook     = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	mimeWorksheet    = "application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"
	mimeStyle        = "application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"
	mimeSharedString = "application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"
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
		el.Formula = c.parsedValue
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
	return encoder.EncodeElement(&el, start)
}

func (p *SheetProtection) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
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

	addr string
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
	z, err := zip.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	return readFile(z)
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

func readFile(z *zip.ReadCloser) (*File, error) {
	if err := readContentFile(z); err != nil {
		return nil, err
	}
	ix := slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == "_rels/.rels"
	})
	if ix < 0 {
		return nil, ErrFile
	}
	wb, err := getWorkbookAddr(z.File[ix])
	if err != nil {
		return nil, err
	}
	ix = slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == wb
	})
	if ix < 0 {
		return nil, ErrFile
	}
	var file *File
	if file, err = getWorkbookSheets(z.File[ix]); err != nil {
		return nil, err
	}
	if err := getWorksheets(file, z); err != nil {
		return nil, err
	}
	return file, nil
}

func readContentFile(z *zip.ReadCloser) error {
	ix := slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == "[Content_Types].xml"
	})
	if ix < 0 {
		return ErrFile
	}
	r, err := z.File[ix].Open()
	if err != nil {
		return err
	}
	_ = r
	return nil
}

func getWorkbookSheets(z *zip.File) (*File, error) {
	r, err := z.Open()
	if err != nil {
		return nil, err
	}
	type xmlSheet struct {
		XMLName xml.Name   `xml:"sheet"`
		Id      string     `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
		Name    string     `xml:"name,attr"`
		Index   int        `xml:"sheetId,attr"`
		State   SheetState `xml:"state,attr"`
	}
	root := struct {
		XMLName xml.Name   `xml:"workbook"`
		Sheets  []xmlSheet `xml:"sheets>sheet"`
	}{}
	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}
	file := NewFile()
	for _, sx := range root.Sheets {
		s := Sheet{
			Id:    sx.Id,
			Name:  sx.Name,
			Index: sx.Index,
			State: sx.State,
		}
		if s.State == 0 {
			s.State = StateVisible
		}
		file.names[s.Name]++
		file.sheets = append(file.sheets, &s)
	}
	return file, nil
}

type relation struct {
	XMLName xml.Name `xml:"Relationship"`
	Target  string   `xml:",attr"`
	Id      string   `xml:",attr"`
	Type    string   `xml:",attr"`
}

func getWorksheets(file *File, z *zip.ReadCloser) error {
	ix := slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == "xl/_rels/workbook.xml.rels"
	})
	if ix < 0 {
		return ErrFile
	}
	r, err := z.File[ix].Open()
	if err != nil {
		return err
	}
	root := struct {
		XMLName   xml.Name   `xml:"Relationships"`
		Relations []relation `xml:"Relationship"`
	}{}
	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return err
	}
	for _, s := range file.sheets {
		ix := slices.IndexFunc(root.Relations, func(r relation) bool {
			return r.Id == s.Id
		})
		if ix < 0 {
			return ErrFile
		}
		s.addr = path.Join("xl", root.Relations[ix].Target)
		if err := getWorksheetData(s, z); err != nil {
			return err
		}
	}
	return nil
}

func getWorksheetData(sheet *Sheet, z *zip.ReadCloser) error {
	ix := slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == sheet.addr
	})
	if ix < 0 {
		return ErrFile
	}
	r, err := z.File[ix].Open()
	if err != nil {
		return err
	}
	root := struct {
		XMLName xml.Name   `xml:"worksheet"`
		Size    *Dimension `xml:"dimension"`
		Rows    []*Row     `xml:"sheetData>row"`
	}{}
	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return err
	}
	sheet.Size = root.Size
	sheet.Rows = root.Rows
	return nil
}

func getWorkbookAddr(z *zip.File) (string, error) {
	r, err := z.Open()
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	root := struct {
		XMLName   xml.Name   `xml:"Relationships"`
		Relations []relation `xml:"Relationship"`
	}{}
	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return "", err
	}
	ix := slices.IndexFunc(root.Relations, func(r relation) bool {
		return strings.HasSuffix(r.Type, "relationships/officeDocument")
	})
	if ix < 0 {
		return "", ErrFile
	}
	return root.Relations[ix].Target, nil
}
