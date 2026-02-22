package eval

import (
	"fmt"

	"github.com/midbel/dockit/grid"
)

type LoaderOptions map[string]any

type Loader interface {
	Open(string) (grid.File, error)
}

type noopLoader struct{}

func (noopLoader) Open(_ string) (grid.File, error) {
	return nil, fmt.Errorf("noop loader can not open file")
}

func defaultLoader() Loader {
	return noopLoader{}
}

type csvLoader struct{}

func CsvLoader() Loader {
	return csvLoader{}
}

func (c csvLoader) Open(file string) (grid.File, error) {
	return nil, nil
}

type xlsxLoader struct{}

func XlsxLoader(ctx *EngineContext) Loader {
	return xlsxLoader{}
}

func (x xlsxLoader) Open(file string) (grid.File, error) {
	return nil, nil
}
