package oxml

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
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

var errFile = errors.New("invalid openxml file")

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
	}{
		Addr: c.Addr(),
		Type: c.Type,
	}
	if c.Type == TypeInlineStr {
		el.InlineValue = c.rawValue
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
	}{}
	if err := decoder.DecodeElement(&el, &start); err != nil {
		return err
	}
	c.Position = parsePosition(el.Addr)
	c.rawValue = el.Value
	c.Type = el.Type

	switch el.Type {
	case TypeInlineStr:
		c.parsedValue = el.Value
	case TypeSharedStr:
	case TypeFormula:
		if !strings.HasPrefix(el.Value, "=") {
			c.Type = TypeInlineStr
		}
	case TypeBool:
		b, _ := strconv.ParseBool(el.Value)
		c.parsedValue = b
	case TypeNumber, "":
		n, _ := strconv.ParseFloat(el.Value, 64)
		c.parsedValue = n
	case TypeDate:
		c.parsedValue = el.Value
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

type Sheet struct {
	Id    string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	Name  string `xml:"name,attr"`
	Index int    `xml:"sheetId,attr"`
	Rows  []*Row
	Size  *Dimension

	Headers []string

	addr string
}

func NewSheet(name string) *Sheet {
	s := Sheet{
		Name: name,
		Size: &Dimension{},
	}
	return &s
}

func (s *Sheet) SetHeaders(headers []string) {
	s.Headers = slices.Clone(headers)
}

func (s *Sheet) Bounding() (Position, Position) {
	var (
		start Position
		end   Position
	)
	start = Position{
		Line:   1,
		Column: 1,
	}
	end = Position{
		Line:   s.Size.Lines,
		Column: s.Size.Columns,
	}
	return start, end
}

func (s *Sheet) Append(data []string) error {
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
	return nil
}

type File struct {
	sheets       []*Sheet
	sharedString []string
}

func NewFile() *File {
	var file File
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

func (f *File) Sheets() []*Sheet {
	return f.sheets
}

// rename a sheet
func (f *File) Rename(oldName, newName string) error {
	return nil
}

// copy a sheet
func (f *File) Copy(oldName, newName string) error {
	return nil
}

func (f *File) AppendSheet(sheet *Sheet) error {
	sheet.Index = len(f.sheets) + 1
	sheet.Id = fmt.Sprintf("rId%d", sheet.Index)
	f.sheets = append(f.sheets, sheet)
	return nil
}

// append sheets of given file to current fule
func (f *File) Merge(other *File) error {
	names := make(map[string]int)
	for _, s := range f.sheets {
		names[s.Name] = 1
	}
	for i, s := range other.sheets {
		s.Index = len(f.sheets) + i + 1
		s.Id = fmt.Sprintf("rId%d", s.Index)
		if n, ok := names[s.Name]; ok {
			names[s.Name] = n + 1
			s.Name = fmt.Sprintf("%s_%03d", s.Name, names[s.Name])
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
		return nil, errFile
	}
	wb, err := getWorkbookAddr(z.File[ix])
	if err != nil {
		return nil, err
	}
	ix = slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == wb
	})
	if ix < 0 {
		return nil, errFile
	}
	var file File
	if file.sheets, err = getWorkbookSheets(z.File[ix]); err != nil {
		return nil, err
	}
	if err := getWorksheets(&file, z); err != nil {
		return nil, err
	}
	return &file, nil
}

func readContentFile(z *zip.ReadCloser) error {
	ix := slices.IndexFunc(z.File, func(f *zip.File) bool {
		return f.Name == "[Content_Types].xml"
	})
	if ix < 0 {
		return errFile
	}
	r, err := z.File[ix].Open()
	if err != nil {
		return err
	}
	_ = r
	// io.Copy(os.Stdout, r)
	return nil
}

func getWorkbookSheets(z *zip.File) ([]*Sheet, error) {
	r, err := z.Open()
	if err != nil {
		return nil, err
	}
	root := struct {
		XMLName xml.Name `xml:"workbook"`
		Sheets  []*Sheet `xml:"sheets>sheet"`
	}{}
	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}
	return root.Sheets, nil
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
		return errFile
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
			return errFile
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
		return errFile
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
		return "", errFile
	}
	return root.Relations[ix].Target, nil
}
