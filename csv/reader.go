package csv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type Reader struct {
	inner         *bufio.Reader
	Comma         byte
	FieldsPerLine int
}

func NewReader(r io.Reader) *Reader {
	rs := Reader{
		inner: bufio.NewReader(r),
		Comma: ',',
	}
	return &rs
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
	line, err := r.inner.ReadBytes('\n')
	if len(line) == 0 && errors.Is(err, io.EOF) {
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
			field, size, err = r.readQuotedField(line[i:])
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
			return line[pos:offset], offset+1, nil
			break
		}
		offset++
	}
	return nil, 0, fmt.Errorf("unterminated quoted field")
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
