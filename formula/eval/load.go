package eval

import (
	"fmt"

	"github.com/midbel/dockit/csv"
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
	fmt.Println(file)
	return csv.Open(file)
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
