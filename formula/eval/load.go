package eval

import (
	"fmt"
	"os"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/oxml"
)

type LoaderOptions map[string]any

type Loader interface {
	Open(string, LoaderOptions) (grid.File, error)
}

type logLoader struct{}

func LogLoader() Loader {
	return logLoader{}
}

func (logLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	pattern, ok := opts["pattern"]
	if !ok {
		return nil, fmt.Errorf("missing pattern to load log file")
	}
	return flat.OpenLog(file, pattern.(string))
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

func (x csvLoader) createReader(file string, opts LoaderOptions) (*csv.Reader, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	rs := csv.NewReader(r)

	if delim, ok := opts["delimiter"]; ok {
		switch delim {
		case "comma", ",":
			rs.Comma = ','
		case "semi", "semicolon", ";":
			rs.Comma = ';'
		case "tab", "\t":
			rs.Comma = '\t'
		case "colon", ":":
			rs.Comma = ':'
		case "detect":
			delim, err := x.detectDelim(file)
			if err != nil {
				return nil, err
			}
			rs.Comma = delim
		default:
			return nil, fmt.Errorf("unsupported csv delimiter %q", delim)
		}
	}
	return rs, nil
}

func (x csvLoader) detectDelim(file string) (byte, error) {
	return csv.Sniff(file)
}

type xlsxLoader struct{}

func XlsxLoader() Loader {
	return xlsxLoader{}
}

func (xlsxLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return oxml.Open(file)
}

type odsLoader struct{}

func OdsLoader() Loader {
	return odsLoader{}
}

func (odsLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return nil, nil
}
