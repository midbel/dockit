package workbook

import (
	"fmt"
	"slices"

	"github.com/midbel/dockit/grid"
)

type Loader interface {
	Name() string
	Detect(string) (bool, error)
	Open(string) (grid.File, error)
}

var registry []Loader

func Register(loader Loader) {
	registry = append(registry, loader)
}

func Open(file string) (grid.File, error) {
	ix := slices.IndexFunc(registry, func(loader Loader) bool {
		ok, err := loader.Detect(file)
		if err != nil {
			ok = false
		}
		return ok
	})
	if ix < 0 {
		return nil, fmt.Errorf("unable to load given file - unsupported format")
	}
	return registry[ix].Open(file)
}
