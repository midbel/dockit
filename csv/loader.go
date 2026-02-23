package csv

import (
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/wbl"
	"path/filepath"
)

type loader struct{}

func NewLoader() wbl.Loader {
	return loader{}
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
