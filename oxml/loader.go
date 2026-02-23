package oxml

import (
	"archive/zip"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/sniff"
	"github.com/midbel/dockit/wbl"
)

type loader struct{}

func NewLoader() wbl.Loader {
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

func (loader) Open(file string) (grid.File, error) {
	return Open(file)
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
