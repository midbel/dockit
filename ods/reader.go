package ods

import (
	"archive/zip"
	"fmt"
	"io"
	"slices"

	sax "github.com/midbel/codecs/xml"
	"github.com/midbel/dockit/grid"
)

const mimeODS = "application/vnd.oasis.opendocument.spreadsheet"

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

const tableNS = "urn:oasis:names:tc:opendocument:xmlns:table:1.0"

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
			_ = sh
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
	sh *Sheet
	reader *sax.Reader
}

func updateSheet(name string, rs *sax.Reader) *sheetReader {
	return &sheetReader{
		sh: NewSheet(name),
		reader: rs,
	}
}

func (r *sheetReader) Update() (*Sheet, error) {
	return r.sh, nil
}
