package csv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

const defaultSheetName = "sheet1"

type Cell struct {
	layout.Position
	raw    string
	parsed value.ScalarValue
}

func (c *Cell) At() layout.Position {
	return c.Position
}

func (c *Cell) Display() string {
	return c.raw
}

func (c *Cell) Value() value.ScalarValue {
	return c.parsed
}

func (c *Cell) Reload(ctx value.Context) error {
	return grid.ErrSupported
}

type row struct {
	Line  int64
	cells []*Cell
}

type Sheet struct {
	rows  []*row
	cells map[layout.Position]*Cell
	size  layout.Dimension
}

func emptySheet() *Sheet {
	return &Sheet{
		cells: make(map[layout.Position]*Cell),
	}
}

func (s *Sheet) Name() string {
	return defaultSheetName
}

func (s *Sheet) Reload(_ value.Context) error {
	return grid.ErrSupported
}

func (s *Sheet) View(rg *layout.Range) grid.View {
	bd := s.Bounds()
	rg.Starts = rg.Starts.Update(bd.Starts)
	rg.Ends = rg.Ends.Update(bd.Ends)
	return grid.NewBoundedView(s, rg)
}

func (s *Sheet) Sub(start, end layout.Position) grid.View {
	return s.View(layout.NewRange(start, end))
}

func (s *Sheet) Bounds() *layout.Range {
	var (
		start layout.Position
		end   layout.Position
	)
	if len(s.rows) == 0 {
		return layout.NewRange(start, end)
	}
	start.Line = 1
	end.Line = int64(len(s.rows))
	if n := len(s.rows[0].cells); n > 0 {
		start.Column = 1
		end.Column = int64(n)
	}
	return layout.NewRange(start, end)
}

func (s *Sheet) Rows() iter.Seq[[]value.ScalarValue] {
	it := func(yield func([]value.ScalarValue) bool) {
		for _, r := range s.rows {
			if len(r.cells) == 0 {
				continue
			}
			res := make([]value.ScalarValue, len(r.cells))
			for i, c := range r.cells {
				res[i] = c.Value()
			}
			if !yield(res) {
				return
			}
		}
	}
	return it
}

func (s *Sheet) Encode(encoder grid.Encoder) error {
	return encoder.EncodeSheet(s)
}

func (s *Sheet) Cell(pos layout.Position) (grid.Cell, error) {
	cell, ok := s.cells[pos]
	if !ok {
		cell = &Cell{
			Position: pos,
			raw:      "",
			parsed:   types.Blank{},
		}
	}
	return cell, nil
}

func (s *Sheet) SetValue(pos layout.Position, val value.ScalarValue) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.raw = val.String()
	c.parsed = val
	return nil
}

func (s *Sheet) SetFormula(_ layout.Position, _ formula.Expr) error {
	return grid.ErrSupported
}

func (s *Sheet) ClearCell(pos layout.Position) error {
	return s.ClearValue(pos)
}

func (s *Sheet) ClearValue(pos layout.Position) error {
	c, ok := s.cells[pos]
	if !ok {
		return grid.NoCell(pos)
	}
	c.raw = ""
	c.parsed = formula.Blank{}
	return nil
}

func (s *Sheet) ClearRange(rg *layout.Range) error {
	return nil
}

func (s *Sheet) ClearFormula(_ layout.Position) error {
	return grid.ErrSupported
}

func (s *Sheet) AppendRow(values []value.ScalarValue) error {
	return nil
}

func (s *Sheet) InsertRow(ix int64, values []value.ScalarValue) error {
	return nil
}

func (s *Sheet) DeleteRow(ix int64) error {
	return nil
}

type File struct {
	sheet *Sheet
}

func NewFile() *File {
	file := File{
		sheet: emptySheet(),
	}
	return &file
}

func Open(file string) (*File, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	sh, err := readSheet(r)
	if err != nil {
		return nil, err
	}
	f := File{
		sheet: sh,
	}
	return &f, nil
}

func (f *File) WriteFile(file string) error {
	w, err := os.Create(file)
	if err != nil {
		return err
	}
	return writeSheet(w, f.sheet)
}

func (f *File) ActiveSheet() (grid.View, error) {
	return f.sheet, nil
}

func (f *File) Sheet(_ string) (grid.View, error) {
	return f.sheet, nil
}

func (f *File) Sheets() []grid.View {
	return []grid.View{f.sheet}
}

func (f *File) Infos() []grid.ViewInfo {
	rg := f.sheet.Bounds()

	i := grid.ViewInfo{
		Name:      f.sheet.Name(),
		Active:    true,
		Protected: false,
		Hidden:    false,
		Size: layout.Dimension{
			Lines:   rg.Height(),
			Columns: rg.Width(),
		},
	}
	return []grid.ViewInfo{i}
}

func (*File) Rename(_, _ string) error {
	return grid.ErrSupported
}

func (*File) Copy(_, _ string) error {
	return grid.ErrSupported
}

func (*File) Remove(_ string) error {
	return grid.ErrSupported
}

func (*File) Reload() error {
	return nil
}

func writeSheet(w io.Writer, sh *Sheet) error {
	ws := NewWriter(w)
	for _, r := range sh.rows {
		row := make([]string, len(r.cells))
		for i := range r.cells {
			row[i] = r.cells[i].Display()
		}
		if err := ws.Write(row); err != nil {
			return err
		}
	}
	ws.Flush()
	return ws.Error()
}

func readSheet(r io.Reader) (*Sheet, error) {
	var (
		rs = NewReader(r)
		sh = emptySheet()
	)
	for line := 1; ; line++ {
		fields, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		r := row{
			Line: int64(line),
		}
		for col, f := range fields {
			p := layout.Position{
				Line:   r.Line,
				Column: int64(col) + 1,
			}
			c := Cell{
				Position: p,
				raw:      f,
				parsed:   formula.Text(f),
			}
			r.cells = append(r.cells, &c)
			sh.cells[p] = &c
		}
		sh.rows = append(sh.rows, &r)
	}
	return sh, nil
}

const (
	quote = '"'
	nl    = '\n'
	cr    = '\r'
	space = ' '
)

var errUnterminated = errors.New("unterminated")

type Reader struct {
	inner         *bufio.Reader
	Comma         byte
	FieldsPerLine int

	atEOF bool
}

func NewReader(r io.Reader) *Reader {
	rs := Reader{
		inner: bufio.NewReader(r),
		Comma: ',',
	}
	return &rs
}

func (r *Reader) Done() bool {
	return r.atEOF
}

func (r *Reader) ReadAll() ([][]string, error) {
	var all [][]string
	for {
		rs, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		all = append(all, rs)
	}
	return all, nil
}

func (r *Reader) Read() ([]string, error) {
	if r.Done() {
		return nil, io.EOF
	}
	line, err := r.inner.ReadBytes(nl)
	if len(line) == 0 && errors.Is(err, io.EOF) {
		r.atEOF = true
		return nil, err
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	var (
		res  []string
		done bool
	)
	for i := 0; i < len(line); {
		var (
			field []byte
			size  int
			err   error
		)
		switch line[i] {
		case cr:
			i++
			if i >= len(line) || line[i] != nl {
				return nil, fmt.Errorf("carriage return only allow followed by newline")
			}
			done = true
		case nl:
			done = true
		case quote:
			for {
				field, size, err = r.readQuotedField(line[i:])
				if err == nil {
					break
				}
				if errors.Is(err, errUnterminated) {
					next, err1 := r.inner.ReadBytes(nl)
					if len(next) == 0 {
						return nil, err
					}
					if err1 != nil && !errors.Is(err1, io.EOF) {
						return nil, err1
					}
					line = append(line, next...)
				} else {
					return nil, err
				}
			}
		default:
			field, size, err = r.readDefaultField(line[i:])
		}
		if done {
			res = append(res, string(field))
			break
		}
		if err != nil {
			return nil, err
		}
		i += size
		if i < len(line) && line[i] != r.Comma && line[i] != cr && line[i] != nl {
			return nil, fmt.Errorf("unexpected character after field")
		}
		i++
		res = append(res, string(field))
	}
	if r.FieldsPerLine > 0 && len(res) != r.FieldsPerLine {
		return nil, fmt.Errorf("invalid number of fields")
	}
	return res, nil
}

func (r *Reader) readQuotedField(line []byte) ([]byte, int, error) {
	var (
		pos    = 1
		offset = pos
	)
	for offset < len(line) {
		if line[offset] == quote {
			if offset+1 < len(line) && line[offset+1] == quote {
				offset += 2
				continue
			}
			return line[pos:offset], offset + 1, nil
			break
		}
		offset++
	}
	return nil, 0, errUnterminated
}

func (r *Reader) readDefaultField(line []byte) ([]byte, int, error) {
	var offset int
	for offset < len(line) {
		switch line[offset] {
		case quote:
			return nil, 0, fmt.Errorf("unexpected quote")
		case r.Comma, cr, nl:
			return line[:offset], offset, nil
		default:
			offset++
		}
	}
	return line[:offset], offset, nil
}

type Writer struct {
	inner *bufio.Writer

	ForceQuote bool
	UseCRLF    bool
	Comma      byte
}

func NewWriter(w io.Writer) *Writer {
	ws := Writer{
		inner: bufio.NewWriter(w),
		Comma: ',',
	}
	return &ws
}

func (w *Writer) WriteAll(data [][]string) error {
	for _, d := range data {
		if err := w.Write(d); err != nil {
			return err
		}
	}
	return w.inner.Flush()
}

func (w *Writer) Write(line []string) error {
	var err error
	for i, str := range line {
		if i > 0 {
			if err = w.inner.WriteByte(w.Comma); err != nil {
				return err
			}
		}
		if w.needQuotes(str) {
			err = w.writeQuoted(str)
		} else {
			_, err = w.inner.WriteString(str)
		}
		if err != nil {
			return err
		}
	}
	if w.UseCRLF {
		err = w.inner.WriteByte(cr)
		if err != nil {
			return err
		}
	}
	err = w.inner.WriteByte(nl)
	return err
}

func (w *Writer) Flush() {
	w.inner.Flush()
}

func (w *Writer) Error() error {
	_, err := w.inner.Write(nil)
	return err
}

func (w *Writer) writeQuoted(str string) error {
	if err := w.inner.WriteByte(quote); err != nil {
		return err
	}
	var err error
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c == quote {
			w.inner.WriteByte(c)
			err = w.inner.WriteByte(c)
		} else if c == cr {
			if w.UseCRLF {
				err = w.inner.WriteByte(c)
			}
		} else if c == nl {
			if w.UseCRLF {
				w.inner.WriteByte(cr)
			}
			err = w.inner.WriteByte(c)
		} else {
			err = w.inner.WriteByte(c)
		}
		if err != nil {
			return err
		}
	}
	err = w.inner.WriteByte(quote)
	return err
}

func (w *Writer) needQuotes(str string) bool {
	if w.ForceQuote {
		return w.ForceQuote
	}
	if str == "" {
		return false
	}
	if str[0] == space {
		return true
	}
	for _, c := range []byte{w.Comma, cr, nl, space} {
		ix := strings.IndexByte(str, c)
		if ix >= 0 {
			return true
		}
	}
	return false
}
