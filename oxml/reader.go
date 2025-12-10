package oxml

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"
)

type reader struct {
	reader *zip.ReadCloser
	base   string

	err error
}

func readFile(name string) (*reader, error) {
	z, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	r := reader{
		reader: z,
		base:   wbBaseDir,
	}
	return &r, nil
}

func (r *reader) Close() error {
	if r.reader == nil {
		return ErrFile
	}
	return r.reader.Close()
}

func (r *reader) ReadFile() (*File, error) {
	file := NewFile()
	r.readContentFile(file)
	// r.readSharedStrings(file)
	r.readWorkbook(file)
	r.readWorksheets(file)
	return file, r.err
}

func (r *reader) readContentFile(file *File) {
	if r.invalid() {
		return
	}
	rs, err := r.openFile("[Content_Types].xml")
	if err != nil {
		r.err = err
		return
	}
	_ = rs
}

func (r *reader) readSharedStrings(file *File) {
	if r.invalid() {
		return
	}
	rs, err := r.openFile(r.fromBase("sharedStrings.xml"))
	if err != nil {
		r.err = err
		return
	}
	_ = rs
}

func (r *reader) readWorkbook(file *File) {
	addr := r.readWorkbookLocation()
	if r.invalid() {
		return
	}
	root := struct {
		XMLName xml.Name   `xml:"workbook"`
		Sheets  []xmlSheet `xml:"sheets>sheet"`
	}{}
	if err := r.decodeXML(addr, &root); err != nil {
		return
	}
	for _, xs := range root.Sheets {
		s := Sheet{
			Id:    xs.Id,
			Name:  xs.Name,
			Index: xs.Index,
			State: xs.State,
		}
		if s.State == 0 {
			s.State = StateVisible
		}
		file.names[s.Name]++
		file.sheets = append(file.sheets, &s)
	}
}

func (r *reader) readWorksheets(file *File) {
	if r.invalid() {
		return
	}
	relations := r.readRelationsForSheets()
	if len(relations) == 0 {
		return
	}
	for _, s := range file.sheets {
		ix := slices.IndexFunc(relations, func(r xmlRelation) bool {
			return r.Id == s.Id
		})
		if ix < 0 {
			r.err = ErrFile
			return
		}
		r.readWorksheet(s, relations[ix].Target)
		if r.invalid() {
			break
		}
	}
}

func (r *reader) readWorksheet(sheet *Sheet, addr string) {
	if r.invalid() {
		return
	}
	root := struct {
		XMLName xml.Name   `xml:"worksheet"`
		Size    *Dimension `xml:"dimension"`
		Rows    []*Row     `xml:"sheetData>row"`
	}{}
	if err := r.decodeXML(r.fromBase(addr), &root); err != nil {
		return
	}
	sheet.Size = root.Size
	sheet.Rows = root.Rows
}

func (r *reader) readWorkbookLocation() string {
	if r.invalid() {
		return ""
	}
	var root xmlRelations
	if err := r.decodeXML("_rels/.rels", &root); err != nil {
		return ""
	}
	ix := slices.IndexFunc(root.Relations, func(r xmlRelation) bool {
		return strings.HasSuffix(r.Type, "relationships/officeDocument")
	})
	if ix < 0 {
		r.err = ErrFile
		return ""
	}
	return root.Relations[ix].Target
}

func (r *reader) readRelationsForSheets() []xmlRelation {
	if r.invalid() {
		return nil
	}
	var root xmlRelations
	if err := r.decodeXML(r.fromBase("_rels/workbook.xml.rels"), &root); err != nil {
		return nil
	}
	return root.Relations
}

func (r *reader) decodeXML(name string, ptr any) error {
	if r.invalid() {
		return r.err
	}
	rs, err := r.openFile(name)
	if err != nil {
		r.err = err
		return r.err
	}
	if err := xml.NewDecoder(rs).Decode(ptr); err != nil {
		r.err = fmt.Errorf("%w: fail to read data from %s", ErrFile, name)
	}
	return r.err
}

func (r *reader) openFile(name string) (io.Reader, error) {
	ix := slices.IndexFunc(r.reader.File, func(f *zip.File) bool {
		return f.Name == name
	})
	if ix < 0 {
		return nil, ErrFile
	}
	return r.reader.File[ix].Open()
}

func (r *reader) fromBase(name string) string {
	parts := append([]string{r.base}, name)
	return strings.Join(parts, "/")
}

func (r *reader) invalid() bool {
	return r.err != nil
}
