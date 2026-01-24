package doc

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/oxml"
)

func Infos(file string) ([]ViewInfo, error) {
	f, err := Open(file)
	if err != nil {
		return nil, err
	}
	return f.Infos(), nil
}

func OpenFormat(file string, format Format) (File, error) {
	switch format {
	case CSV:
		return csv.Open(file)
	case OXML:
		return oxml.Open(file)
	case ODS:
	default:
	}
	return nil, fmt.Errorf("unsupported format")
}

func Open(file string) (File, error) {
	format, err := detectFormat(file)
	if err != nil {
		return nil, err
	}
	return OpenFormat(file, format)
}

type Format int

const (
	CSV Format = iota << 1
	OXML
	ODS
	// Delimited
	// Json
	// Xml
	Unknown
)

func detectFormat(file string) (Format, error) {
	if ok, err := isZip(file); ok && err == nil {
		return detectZip(file)
	}
	return CSV, nil
}

func detectZip(file string) (Format, error) {
	z, err := zip.OpenReader(file)
	if err != nil {
		return Unknown, err
	}
	for _, f := range z.File {
		switch f.Name {
		case "xl/workbook.xml", "[Content_Types].xml":
			return OXML, nil
		case "mimetype":
			r, _ := f.Open()
			defer r.Close()

			buf, _ := io.ReadAll(r)
			if string(buf) == "application/vnd.oasis.opendocument.spreadsheet" {
				return ODS, nil
			}
		default:
		}
	}
	return Unknown, nil
}

var magicZipBytes = [][]byte{
	{0x50, 0x4b, 0x03, 0x04},
	{0x50, 0x4b, 0x05, 0x06},
	{0x50, 0x4b, 0x07, 0x08},
}

func isZip(file string) (bool, error) {
	r, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer r.Close()

	magic := make([]byte, 4)
	if n, err := io.ReadFull(r, magic); err != nil || n != len(magic) {
		return false, err
	}
	for _, mzb := range magicZipBytes {
		if bytes.Equal(magic, mzb) {
			return true, nil
		}
	}
	return false, nil
}
