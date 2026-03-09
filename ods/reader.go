package ods

import (
	"archive/zip"
	"fmt"
	"io"
	"slices"
	"strconv"

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
	rx.Element(qn, func(rs *sax.Reader, e sax.E) error {
		sr := updateSheet(e.GetAttributeValue("name"), rs)
		sh, err := sr.Update()
		if err == nil {
			file.sheets = append(file.sheets, sh)
		}
		return err
	})
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

type sheetReader struct {
	sh     *Sheet
	reader *sax.Reader
}

func updateSheet(name string, rs *sax.Reader) *sheetReader {
	return &sheetReader{
		sh:     NewSheet(name),
		reader: rs,
	}
}

func (r *sheetReader) Update() (*Sheet, error) {
	var (
		qCell = sax.ExpandedName("table-cell", "table", tableNS)
		qRow  = sax.ExpandedName("table-row", "table", tableNS)
	)
	r.reader.Element(qRow, r.onRow)
	r.reader.Element(qCell, r.onCell)
	return r.sh, nil
}

func (r *sheetReader) onRow(_ *sax.Reader, e sax.E) error {
	var rs row
	r.sh.rows = append(r.sh.rows, &rs)
	r.sh.Size.Lines++
	return nil
}

func (r *sheetReader) onCell(rs *sax.Reader, e sax.E) error {
	if len(r.sh.rows) == 0 {
		return fmt.Errorf("no row in worksheet")
	}
	var (
		pos  = len(r.sh.rows) - 1
		cell = &Cell{
			Position: layout.Position{
				Line:   int64(pos + 1),
				Column: int64(len(r.sh.rows[pos].Cells)) + 1,
			},
		}
	)
	r.sh.rows[pos].Cells = append(r.sh.rows[pos].Cells, cell)
	r.sh.cells[cell.Position] = cell

	switch e.GetAttributeValue("value-type") {
	case "float", "currency", "percentage":
		cell.raw = e.GetAttributeValue("value")
		val, err := strconv.ParseFloat(cell.raw, 64)
		if err != nil {
			return err
		}
		cell.parsed = value.Float(val)
	case "boolean":
		cell.raw = e.GetAttributeValue("boolean-value")
		val, err := strconv.ParseBool(cell.raw)
		if err != nil {
			return err
		}
		cell.parsed = value.Boolean(val)
	case "date":
		cell.raw = e.GetAttributeValue("date-value")
	case "time":
		cell.raw = e.GetAttributeValue("time-value")
	default:
	}
	if cell.raw != "" {
		return nil
	}
	qn := sax.ExpandedName("p", "text", textNS)
	rs.Element(qn, func(rs *sax.Reader, _ sax.E) error {
		rs.OnText(func(_ *sax.Reader, str string) error {
			cell.raw = str
			cell.parsed = value.Text(str)
			return nil
		})
		return nil
	})
	return nil
}
