package csv

import (
	"bufio"
	"io"
)

type Reader struct {
	inner         *bufio.Reader
	Comma         rune
	FieldsPerLine int
	TrimSpace     bool
}

func NewReader(r io.Reader) *Reader {
	rs := Reader{
		inner: r,
		Comma: ',',
	}
	return &rs
}

func (r *Reader) Read() ([]string, error) {
	return nil, nil
}

func (r *Reader) ReadAll() ([][]string, error) {
	return nil, nil
}
