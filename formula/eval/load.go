package eval

import (
	"fmt"
	"io"
	"os"

	"github.com/midbel/codecs/json"
	"github.com/midbel/codecs/xml"
	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/ods"
	"github.com/midbel/dockit/oxml"
	"github.com/midbel/dockit/value"
	"github.com/midbel/probe"
)

type LoaderOptions map[string]any

type Loader interface {
	Open(string, LoaderOptions) (grid.File, error)
}

func (o LoaderOptions) getAsString(key string) string {
	v, ok := o[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

type Writer interface {
	Write(string, grid.File) error
}

type logLoader struct{}

func LogLoader() Loader {
	return logLoader{}
}

func (logLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	pattern := opts.getAsString("pattern")
	return flat.OpenLog(file, pattern)
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

	switch delim := opts.getAsString("delimiter"); delim {
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
	return rs, nil
}

func (c csvLoader) detectDelim(file string) (byte, error) {
	return csv.Sniff(file)
}

type structuredLoader struct {
	decoder func(r io.Reader) (any, error)
}

func JsonLoader() Loader {
	return structuredLoader{
		decoder: json.Decode,
	}
}

func Json5Loader() Loader {
	return structuredLoader{
		decoder: json.Decode5,
	}
}

func XmlLoader() Loader {
	return structuredLoader{
		decoder: xml.Decode,
	}
}

func (j structuredLoader) Open(file string, opts LoaderOptions) (grid.File, error) {
	result, err := j.readFile(file, opts)
	if err != nil {
		return nil, err
	}
	var sheets []*flat.Sheet
	for i, set := range result.Sets {
		arr, ok := set.([]any)
		if !ok {
			return nil, fmt.Errorf("expected array")
		}
		values, err := j.processData(arr)
		if err != nil {
			return nil, err
		}
		sh := flat.NewSheet(fmt.Sprintf("sheet%d", i+1), values)
		sheets = append(sheets, sh)
	}
	return flat.NewFileFromSheets(sheets...), nil
}

func (structuredLoader) processData(arr []any) ([][]value.ScalarValue, error) {
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
			case nil:
				x = value.ErrNA
			default:
				return nil, fmt.Errorf("unexpected type from json array")
			}
			row = append(row, x)
		}
		values = append(values, row)
	}
	return values, nil
}

func (j structuredLoader) readFile(file string, opts LoaderOptions) (*probe.Result, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := j.decoder(r)
	if err != nil {
		return nil, err
	}
	var (
		ps  probe.Options
		val string
	)
	val = opts.getAsString("missing")
	if ps.Missing, err = probe.ParseMissingMode(val); val != "" && err != nil {
		return nil, err
	}
	val = opts.getAsString("zip")
	if ps.Zip, err = probe.ParseZipMode(val); val != "" && err != nil {
		return nil, err
	}
	val = opts.getAsString("expand")
	if ps.Expand, err = probe.ParseExpandMode(val); val != "" && err != nil {
		return nil, err
	}
	query := opts.getAsString("query")
	return probe.Execute(query, data, &ps)
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
