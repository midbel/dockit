package oxml

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"path"
	"slices"
	"strconv"
	"strings"
	"io"
	"os"
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

var errFile = errors.New("invalid openxml file")

type Cell struct {
	rawValue    string
	parsedValue any
	Position
}

func (c *Cell) Value() string {
	return c.rawValue
}

func (c *Cell) Get() any {
	return c.parsedValue
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

	switch el.Type {
	case TypeInlineStr:
		c.parsedValue = el.Value
	case TypeSharedStr:
	case TypeFormula:
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

	addr  string
}

type File struct {
	sheets []*Sheet
	sharedString []string
}

func Open(file string) (*File, error) {
	z, err := zip.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	return readFile(z)
}

func (f *File) SheetNames() []string {
	var sheets []string
	for _, s := range f.sheets {
		sheets = append(sheets, s.Name)
	}
	return sheets
}

// rename a sheet
func (f *File) Rename(oldName, newName string) error {
	return nil
}

// copy a sheet
func (f *File) Copy(oldName, newName string) error {
	return nil
}

func (f *File) Merge(other *File) error {
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
	io.Copy(os.Stdout, r)
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
