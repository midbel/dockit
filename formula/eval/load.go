package eval

import (
	"fmt"
	"os"

	"github.com/midbel/codecs/json"
	"github.com/midbel/codecs/probe"
	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/ods"
	"github.com/midbel/dockit/oxml"
	"github.com/midbel/dockit/value"
)

type LoaderOptions map[string]any

type Loader interface {
	Open(string, LoaderOptions) (grid.File, error)
}

type Writer interface {
	Write(string, grid.File) error
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

func (c csvLoader) Write(out string, file grid.File) error {
	return nil
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
			delim, err := c.detectDelim(file)
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

func (c csvLoader) detectDelim(file string) (byte, error) {
	return csv.Sniff(file)
}

type jsonLoader struct {
	five bool
}

func JsonLoader() Loader {
	return jsonLoader{}
}

func Json5Loader() Loader {
	return jsonLoader{
		five: true,
	}
}

func (j jsonLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	arr, err := j.readFile(file)
	if err != nil {
		return nil, err
	}
	values, err := j.processData(arr)
	if err != nil {
		return nil, err
	}
	return flat.NewFileFromRows(values), nil
}

func (jsonLoader) processData(arr []any) ([][]value.ScalarValue, error) {
	var values [][]value.ScalarValue
	for i := range arr {
		vs, ok := arr[i].([]any)
		if !ok {
			return nil, fmt.Errorf("array expected")
		}
		row := make([]value.ScalarValue, 0, len(vs))
		for j := range vs {
			var x value.ScalarValue
			switch v := vs[j].(type) {
			case float64:
				x = value.Float(v)
			case string:
				x = value.Text(v)
			case bool:
				x = value.Boolean(v)
			default:
				return nil, fmt.Errorf("unexpected type from json array")
			}
			row = append(row, x)
		}
		values = append(values, row)
	}
	return values, nil
}

func (jsonLoader) readFile(file string) ([]any, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := json.Decode(r)
	if err != nil {
		return nil, err
	}
	result, err := probe.Traverse("", data, nil)
	if err != nil {
		return nil, err
	}
	arr, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("array expected")
	}
	return arr, nil
}

type xmlLoader struct{}

func XmlLoader() Loader {
	return xmlLoader{}
}

func (xmlLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return nil, nil
}

type xlsxLoader struct{}

func XlsxLoader() Loader {
	return xlsxLoader{}
}

func (x xlsxLoader) Write(out string, file grid.File) error {
	w, err := os.Create(out)
	if err != nil {
		return err
	}
	defer w.Close()
	return nil
}

func (xlsxLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return oxml.Open(file)
}

type odsLoader struct{}

func OdsLoader() Loader {
	return odsLoader{}
}

func (odsLoader) Write(out string, file grid.File) error {
	w, err := os.Create(out)
	if err != nil {
		return err
	}
	defer w.Close()
	return nil
}

func (odsLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	return ods.Open(file)
}
