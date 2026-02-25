package csv

import (
	"path/filepath"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/workbook"
)

type loader struct {
	sep byte
}

func NewCommaLoader() workbook.Loader {
	return createLoader(',')
}

func NewTabLoader() workbook.Loader {
	return createLoader('\t')
}

func NewSemicolonLoader() workbook.Loader {
	return createLoader(';')
}

func NewColonLoader() workbook.Loader {
	return createLoader(':')
}

func createLoader(sep byte) workbook.Loader {
	return loader{
		sep: sep,
	}
}

func (loader) Name() string {
	return "csv"
}

func (loader) Detect(file string) (bool, error) {
	return filepath.Ext(file) == ".csv", nil
}

func (loader) Open(file string) (grid.File, error) {
	return Open(file)
}
