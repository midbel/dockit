package ods

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/midbel/dockit/driver"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/sniff"
)

type loader struct{}

func NewLoader() driver.Loader {
	return loader{}
}

func (loader) Name() string {
	return "ods"
}

func (loader) Detect(file string) (bool, error) {
	ok, err := sniff.IsZip(file)
	if err != nil || !ok {
		return ok, err
	}
	return isOds(file)
}

func (loader) New() (grid.File, error) {
	return nil, nil
}

func (loader) Open(file string) (grid.File, error) {
	return Open(file)
}

func (loader) IsSupportedExt(ext string) bool {
	return ext == ".ods"
}

func isOds(file string) (bool, error) {
	z, err := zip.OpenReader(file)
	if err != nil {
		return false, err
	}
	if len(z.File) == 0 {
		return false, fmt.Errorf("empty zip file")
	}
	if z.File[0].Name != "mimetype" {
		return false, nil
	}
	rs, err := z.File[0].Open()
	if err != nil {
		return false, err
	}
	buf, _ := io.ReadAll(rs)
	return string(buf) == mimeODS, nil
}
