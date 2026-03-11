package ods

import (
	"archive/zip"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	sax "github.com/midbel/codecs/xml"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

const mimeODS = "application/vnd.oasis.opendocument.spreadsheet"

const (
	tableNS = "urn:oasis:names:tc:opendocument:xmlns:table:1.0"
	textNS  = "urn:oasis:names:tc:opendocument:xmlns:text:1.0"
)

type reader struct {
	reader *zip.ReadCloser
	err    error
}

func readFile(name string) (*reader, error) {
	z, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	r := reader{
		reader: z,
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
	r.readMime(file)
	r.readSettings(file)
	r.readMeta(file)
	r.readStyle(file)
	r.readContent(file)
	return file, r.err
}

func (r *reader) readMime(file *File) {
	if r.invalid() {
		return
	}
	rs, err := r.openFile("mimetype")
	if err != nil {
		r.err = err
		return
	}
	buf, err := io.ReadAll(rs)
	if err != nil {
		r.err = err
		return
	}
	if string(buf) != mimeODS {
		r.err = fmt.Errorf("%w: invalid mimetype", grid.ErrFile)
	}
}

func (r *reader) readSettings(file *File) error {
	return nil
}

func (r *reader) readMeta(file *File) error {
	return nil
}

func (r *reader) readStyle(file *File) error {
	return nil
}

func (r *reader) readContent(file *File) {
	if r.invalid() {
		return
	}
	rs, err := r.openFile("content.xml")
	if err != nil {
		r.err = err
		return
	}
	var (
		rx = sax.NewReader(rs)
		qn = sax.ExpandedName("table", "table", tableNS)
	)
	rx.HandleElement(qn, handleTable(file))
	r.err = rx.Start()
}

func (r *reader) openFile(name string) (io.Reader, error) {
	ix := slices.IndexFunc(r.reader.File, func(f *zip.File) bool {
		return f.Name == name
	})
	if ix < 0 {
		return nil, fmt.Errorf("%w: file %s not found in archive", grid.ErrFile, name)
	}
	return r.reader.File[ix].Open()
}

func (r *reader) invalid() bool {
	return r.err != nil
}

type tableHandler struct {
	sheet *Sheet
	file  *File
}

func handleTable(file *File) *tableHandler {
	h := &tableHandler{
		file: file,
	}
	h.Reset()
	return h
}

func (h *tableHandler) Reset() {
	h.sheet = NewSheet("")
}

func (h *tableHandler) Open(rs *sax.Reader, e sax.E) error {
	h.sheet.Label = e.GetAttributeValue("name")
	if vz := e.GetAttributeValue("display"); vz == "" || vz == "true" {
		h.sheet.Visible = true
	}
	if lk := e.GetAttributeValue("protected"); lk == "" || lk == "true" {
		h.sheet.Locked = true
	}

	qn := sax.ExpandedName("table-row", "table", tableNS)
	rs.HandleElement(qn, handleRow(h.sheet))
	return nil
}

func (h *tableHandler) Close(rs *sax.Reader, _ sax.E) error {
	defer h.Reset()
	h.file.sheets = append(h.file.sheets, h.sheet)
	return nil
}

type rowHandler struct {
	line   int
	repeat int
	sheet  *Sheet

	cell *cellHandler
}

func handleRow(sh *Sheet) *rowHandler {
	h := &rowHandler{
		sheet: sh,
		cell:  handleCell(sh),
	}
	h.Reset()
	return h
}

func (h *rowHandler) Reset() {
	h.repeat = 1
}

func (h *rowHandler) Open(rs *sax.Reader, e sax.E) error {
	h.cell.Reset()

	var (
		val   = e.GetAttributeValue("number-rows-repeated")
		count = 1
	)
	if val != "" {
		c, err := strconv.Atoi(val)
		if err != nil || c < 1 {
			return fmt.Errorf("%w: invalid value number-rows-repeated", grid.ErrFile)
		}
		count = c
	}
	h.repeat = count
	h.line++
	row := &row{
		Line: int64(h.line),
	}
	h.sheet.rows = append(h.sheet.rows, row)

	qn := sax.ExpandedName("table-cell", "table", tableNS)
	rs.HandleElement(qn, h.cell)
	return nil
}

func (h *rowHandler) Close(rs *sax.Reader, e sax.E) error {
	defer h.Reset()

	pos := len(h.sheet.rows)
	h.sheet.Size.Lines += int64(h.repeat)
	h.sheet.Size.Columns = max(h.sheet.Size.Columns, int64(len(h.sheet.rows[pos-1].Cells)))

	if h.repeat == 1 {
		return nil
	}
	h.line += h.repeat - 1

	curr := h.sheet.rows[pos-1]
	if len(curr.Cells) == 0 {
		return nil
	}
	for i := 1; i < h.repeat; i++ {
		rs := curr.Clone()
		for j := range rs.Cells {
			rs.Cells[j].Line += int64(i)
		}
		h.sheet.rows = append(h.sheet.rows, rs)

	}
	return nil
}

type cellHandler struct {
	column int

	repeat int
	raw    string
	parsed value.ScalarValue

	sheet *Sheet
	text  *textHandler
}

func handleCell(sh *Sheet) *cellHandler {
	h := &cellHandler{
		sheet: sh,
		text:  new(textHandler),
	}
	h.Reset()
	return h
}

func (h *cellHandler) Reset() {
	h.repeat = 1
	h.column = 0
	h.raw = ""
	h.parsed = nil
	h.text.Reset()
}

func (h *cellHandler) Open(rs *sax.Reader, e sax.E) error {
	if len(h.sheet.rows) == 0 {
		return fmt.Errorf("%w: sheet is empty", grid.ErrFile)
	}

	var (
		val   = e.GetAttributeValue("number-columns-repeated")
		count = 1
	)
	if val != "" {
		c, err := strconv.Atoi(val)
		if err != nil || c < 1 {
			return fmt.Errorf("%w: invalid value number-columns-repeated", grid.ErrFile)
		}
		count = c
	}
	h.repeat = count

	switch e.GetAttributeValue("value-type") {
	case "float", "currency", "percentage":
		h.raw = e.GetAttributeValue("value")
		val, err := strconv.ParseFloat(h.raw, 64)
		if err != nil {
			return err
		}
		h.parsed = value.Float(val)
	case "boolean":
		h.raw = e.GetAttributeValue("boolean-value")
		val, err := strconv.ParseBool(h.raw)
		if err != nil {
			return err
		}
		h.parsed = value.Boolean(val)
	case "date":
		h.raw = e.GetAttributeValue("date-value")
		val, err := time.Parse("2006-01-02", h.raw)
		if err != nil {
			return err
		}
		h.parsed = value.Date(val)
	case "time":
		h.raw = e.GetAttributeValue("time-value")
	default:
	}
	if h.raw != "" {
		return nil
	}
	qn := sax.ExpandedName("p", "text", textNS)
	rs.HandleElement(qn, h.text)
	return nil
}

func (h *cellHandler) Close(rs *sax.Reader, e sax.E) error {
	defer h.Reset()

	if h.parsed == nil && h.text.Empty() {
		h.column += h.repeat
		return nil
	}
	var (
		ix  = len(h.sheet.rows)
		row = h.sheet.rows[ix-1]
	)
	for i := 0; i < h.repeat; i++ {
		h.column++
		cell := h.create(int64(ix))
		row.Cells = append(row.Cells, cell)
		h.sheet.cells[cell.Position] = cell
	}

	return nil
}

func (h *cellHandler) create(line int64) *Cell {
	cell := Cell{
		raw:      h.raw,
		parsed:   h.parsed,
		Position: layout.NewPosition(line, int64(h.column)),
	}
	if cell.parsed == nil {
		cell.raw = h.text.Raw()
		cell.parsed = h.text.Value()
	}
	return &cell
}

type textHandler struct {
	parts []string
}

func (h *textHandler) Reset() {
	h.parts = h.parts[:0]
}

func (h *textHandler) Raw() string {
	return strings.Join(h.parts, "\n")
}

func (h *textHandler) Value() value.ScalarValue {
	return value.Text(h.Raw())
}

func (h *textHandler) Empty() bool {
	return len(h.parts) == 0
}

func (h *textHandler) Open(rs *sax.Reader, el sax.E) error {
	rs.OnText(func(_ *sax.Reader, str string) error {
		h.parts = append(h.parts, str)
		return nil
	})
	return nil
}

func (h *textHandler) Close(rs *sax.Reader, el sax.E) error {
	return nil
}
