package oxml

import (
	"archive/zip"

	"github.com/midbel/dockit/driver"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/sniff"
)

type loader struct{}

func NewLoader() driver.Loader {
	return loader{}
}

func (loader) Name() string {
	return "openxml"
}

func (loader) Detect(file string) (bool, error) {
	ok, err := sniff.IsZip(file)
	if err != nil || !ok {
		return ok, err
	}
	return isOxml(file)
}

func (loader) New() (grid.File, error) {
	return NewFile(), nil
}

func (loader) Open(file string) (grid.File, error) {
	return Open(file)
}

func (loader) IsSupportedExt(ext string) bool {
	return ext == ".xlsx"
}

func isOxml(file string) (bool, error) {
	z, err := zip.OpenReader(file)
	if err != nil {
		return false, err
	}
	for _, f := range z.File {
		switch f.Name {
		case "xl/workbook.xml", "[Content_Types].xml":
			return true, nil
		default:
		}
	}
	return false, nil
}
