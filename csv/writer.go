package csv

import (
	"bufio"
	"io"
	"strings"
)

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

func (w *Writer) Err() error {
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
