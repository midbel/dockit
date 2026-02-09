package oxml

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	sax "github.com/midbel/codecs/xml"
	"github.com/midbel/dockit/formula/eval"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
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
		return grid.ErrFile
	}
	return r.reader.Close()
}

func (r *reader) ReadFile() (*File, error) {
	file := NewFile()
	r.readContentFile(file)
	r.readSharedStrings(file)
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
	root := struct {
		XMLName xml.Name `xml:"sst"`
		Shared  []string `xml:"si>t"`
	}{}
	if err := r.decodeXML(r.fromBase("sharedStrings.xml"), &root); err != nil {
		r.err = nil
		return
	}
	file.sharedStrings = root.Shared
}

func (r *reader) readWorkbook(file *File) {
	addr := r.readWorkbookLocation()
	if r.invalid() {
		return
	}
	var root xmlWorkbook
	if err := r.decodeXML(addr, &root); err != nil {
		return
	}
	for i, xs := range root.Sheets {
		s := Sheet{
			Id:    xs.Id,
			Label: xs.Name,
			Index: xs.Index,
			State: xs.State,
			cells: make(map[layout.Position]*Cell),
		}
		if i == root.View.ActiveTab {
			s.Active = true
		}
		if s.State == 0 {
			s.State = StateVisible
		}
		file.names[s.Label]++
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
			r.err = grid.ErrFile
			return
		}
		r.readWorksheet(s, file.sharedStrings, relations[ix].Target)
		if r.invalid() {
			break
		}
	}
}

func (r *reader) readWorksheet(sheet *Sheet, sharedStrings []string, addr string) {
	if r.invalid() {
		return
	}
	z, err := r.openFile(r.fromBase(addr))
	if err != nil {
		r.err = err
		return
	}
	rs := updateSheet(z, sheet, sharedStrings)
	if err := rs.Update(); err != nil {
		r.err = err
	}
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
		r.err = grid.ErrFile
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
		r.err = fmt.Errorf("%w: fail to read data from %s", grid.ErrFile, name)
	}
	return r.err
}

func (r *reader) openFile(name string) (io.Reader, error) {
	ix := slices.IndexFunc(r.reader.File, func(f *zip.File) bool {
		return f.Name == name
	})
	if ix < 0 {
		return nil, grid.ErrFile
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

type sharedFormula struct {
	layout.Position
	Expr value.Formula
}

type sheetReader struct {
	reader         *sax.Reader
	sheet          *Sheet
	sharedStrings  []string
	sharedFormulas map[string]sharedFormula
}

func updateSheet(r io.Reader, sheet *Sheet, shared []string) *sheetReader {
	if sheet.cells == nil {
		sheet.cells = make(map[layout.Position]*Cell)
	}
	rs := sheetReader{
		reader:         sax.NewReader(r),
		sheet:          sheet,
		sharedStrings:  shared,
		sharedFormulas: make(map[string]sharedFormula),
	}
	return &rs
}

func (r *sheetReader) Update() error {
	r.reader.Element(sax.LocalName("dimension"), r.onDimension)
	r.reader.Element(sax.LocalName("sheetProtection"), r.onProtection)
	r.reader.Element(sax.LocalName("row"), r.onRow)
	r.reader.Element(sax.LocalName("c"), r.onCell)
	return r.reader.Start()
}

func (r *sheetReader) parseCellValue(cell *Cell, str string) error {
	cell.raw = str
	switch cell.Type {
	case TypeSharedStr:
		n, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("invalid shared string index: %s", str)
		}
		if n < 0 || n >= len(r.sharedStrings) {
			return fmt.Errorf("shared string index out of bounds")
		}
		cell.parsed = value.Text(r.sharedStrings[n])
	case TypeDate:
		// date: TBW
	case TypeInlineStr:
		cell.parsed = value.Text(str)
	case TypeBool:
		b, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		cell.parsed = value.Boolean(b)
	default:
		n, err := strconv.ParseFloat(strings.TrimSpace(str), 64)
		if err != nil {
			cell.parsed = value.Text(str)
		} else {
			cell.parsed = value.Float(n)
		}
	}
	return nil
}

func (r *sheetReader) parseCellFormula(cell *Cell, el sax.E, rs *sax.Reader) error {
	var (
		shared = el.GetAttributeValue("t")
		index  = el.GetAttributeValue("si")
	)
	if sf, ok := r.sharedFormulas[index]; shared == "shared" && ok {
		pos := layout.Position{
			Line:   cell.Line - sf.Line,
			Column: cell.Column - sf.Column,
		}
		cell.formula = eval.CloneWithOffset(sf.Expr, pos)
	}
	if el.SelfClosed {
		return nil
	}
	rs.OnText(func(_ *sax.Reader, str string) error {
		formula, err := eval.ParseFormula(str)
		if err != nil {
			return err
		}
		if _, ok := r.sharedFormulas[index]; shared == "shared" && !ok {
			r.sharedFormulas[index] = sharedFormula{
				Position: cell.Position,
				Expr:     formula,
			}
		}
		if cell.formula == nil {
			cell.formula = formula
		}
		return nil
	})
	return nil
}

func (r *sheetReader) onCell(rs *sax.Reader, el sax.E) error {
	if len(r.sheet.rows) == 0 {
		return fmt.Errorf("no row in worksheet")
	}

	var (
		kind  = el.GetAttributeValue("t")
		index = el.GetAttributeValue("r")
		local sax.QName
		pos   = len(r.sheet.rows) - 1
		cell  = &Cell{
			Position: layout.ParsePosition(index),
			Type:     kind,
			dirty:    true,
		}
	)
	if kind == TypeInlineStr {
		local = sax.LocalName("is")
	} else {
		local = sax.LocalName("v")
	}
	r.sheet.rows[pos].Cells = append(r.sheet.rows[pos].Cells, cell)
	r.sheet.cells[cell.Position] = cell

	rs.Element(local, func(rs *sax.Reader, _ sax.E) error {
		rs.OnText(func(_ *sax.Reader, str string) error {
			return r.parseCellValue(cell, str)
		})
		return nil
	})
	rs.Element(sax.LocalName("f"), func(rs *sax.Reader, el sax.E) error {
		return r.parseCellFormula(cell, el, rs)
	})
	return nil
}

func (r *sheetReader) onRow(rs *sax.Reader, el sax.E) error {
	var (
		oxr row
		err error
	)
	oxr.Line, err = strconv.ParseInt(el.GetAttributeValue("r"), 10, 64)
	oxr.Hidden = el.GetAttributeValue("hidden") == "1"
	if err == nil {
		r.sheet.rows = append(r.sheet.rows, &oxr)
	}
	return err
}

func (r *sheetReader) onDimension(rs *sax.Reader, el sax.E) error {
	startIx, endIx, ok := strings.Cut(el.GetAttributeValue("ref"), ":")
	if ok {
		var (
			start = layout.ParsePosition(startIx)
			end   = layout.ParsePosition(endIx)
		)
		r.sheet.Size.Lines = (end.Line - start.Line) + 1
		r.sheet.Size.Columns = (end.Column - start.Column) + 1
	} else {
		r.sheet.Size.Lines = 1
		r.sheet.Size.Columns = 1
	}
	return nil
}

func (r *sheetReader) onProtection(rs *sax.Reader, el sax.E) error {
	if el.GetAttributeValue("sheet") == "1" {
		r.sheet.Protected |= ProtectedSheet
	}
	if el.GetAttributeValue("objects") == "1" {
		r.sheet.Protected |= ProtectedObjects
	}
	if el.GetAttributeValue("scenarios") == "1" {
		r.sheet.Protected |= ProtectedScenarios
	}
	if el.GetAttributeValue("formatCells") == "1" {
		r.sheet.Protected |= ProtectedFormatCells
	}
	if el.GetAttributeValue("formatColumns") == "1" {
		r.sheet.Protected |= ProtectedFormatColumns
	}
	if el.GetAttributeValue("formatRows") == "1" {
		r.sheet.Protected |= ProtectedFormatRows
	}
	if el.GetAttributeValue("deleteColumns") == "1" {
		r.sheet.Protected |= ProtectedDeleteColumns
	}
	if el.GetAttributeValue("deleteRows") == "1" {
		r.sheet.Protected |= ProtectedDeleteRows
	}
	if el.GetAttributeValue("insertColumns") == "1" {
		r.sheet.Protected |= ProtectedInsertColumns
	}
	if el.GetAttributeValue("insertRows") == "1" {
		r.sheet.Protected |= ProtectedInsertRows
	}
	if el.GetAttributeValue("sort") == "1" {
		r.sheet.Protected |= ProtectedSort
	}
	return nil
}
