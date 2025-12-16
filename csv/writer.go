package csv

import (
	"bufio"
	"io"
	"strings"
	"unicode/utf8"
)

type Writer struct {
	inner *bufio.Writer

	ForceQuote bool
	UseCRLF    bool
	Comma      rune
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
			if _, err = w.inner.WriteRune(w.Comma); err != nil {
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
		_, err = w.inner.WriteRune(cr)
		if err != nil {
			return err
		}
	}
	_, err = w.inner.WriteRune(nl)
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
	if _, err := w.inner.WriteRune(quote); err != nil {
		return err
	}
	var err error
	for i := 0; i < len(str); {
		c, z := utf8.DecodeRuneInString(str[i:])
		if c == utf8.RuneError {
			break
		}
		if c == quote {
			w.inner.WriteRune(c)
			_, err = w.inner.WriteRune(c)
		} else if c == cr {
			if w.UseCRLF {
				_, err = w.inner.WriteRune(c)
			}
		} else if c == nl {
			if w.UseCRLF {
				w.inner.WriteRune(cr)
			}
			_, err = w.inner.WriteRune(c)
		} else {
			_, err = w.inner.WriteRune(c)
		}
		if err != nil {
			return err
		}
		i += z
	}
	_, err = w.inner.WriteRune(quote)
	return err
}

func (w *Writer) needQuotes(str string) bool {
	if w.ForceQuote {
		return w.ForceQuote
	}
	if str == "" {
		return false
	}
	if r, _ := utf8.DecodeRuneInString(str); r == space {
		return true
	}
	if w.Comma < utf8.RuneSelf {
		for i := 0; i < len(str); i++ {
			c := str[i]
			if c == nl || c == cr || c == quote || c == byte(w.Comma) {
				return true
			}
		}
	} else {
		ok := strings.ContainsRune(str, w.Comma) || strings.ContainsAny(str, "\"\r\n")
		if ok {
			return ok
		}
	}
	return false
}
