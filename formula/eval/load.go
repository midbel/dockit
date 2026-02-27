package eval

import (
	"fmt"

	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/oxml"
)

type LoaderOptions map[string]any

type Loader interface {
	Open(string, LoaderOptions) (grid.File, error)
}

type csvLoader struct{}

func CsvLoader() Loader {
	return csvLoader{}
}

func (c csvLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	rs, err := c.createReader(file, opts)
	if err != nil {
		return nil, err
	}
	return flat.OpenReader(rs)
}

func (c csvLoader) createReader(file string, opts LoaderOptions) (*csv.Reader, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	rs := csv.NewReader(r)

	if delim, ok := opts["delimiter"]; ok {
		switch delimiter {
		case "comma", ",":
			rs.Comma = ','
		case "semi", "semicolon", ";":
			rs.Comma = ';'
		case "tab", "\t":
			rs.Comma = '\t'
		case "colon", ":":
			rs.Comma = ':'
		default:
			return nil, fmt.Errorf("unsupported csv delimiter %q", delim)
		}
	}

	return rs, nil
}

type xlsxLoader struct{}

func XlsxLoader() Loader {
	return xlsxLoader{}
}

func (x xlsxLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return oxml.Open(file)
}

type odsLoader struct{}

func OdsLoader() Loader {
	return odsLoader{}
}

func (x odsLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return nil, nil
}

type logLoader struct{}

func LogLoader() Loader {
	return logLoader{}
}

func (g logLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return nil, nil
}
