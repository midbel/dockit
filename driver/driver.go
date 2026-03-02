package driver

import (
	"github.com/midbel/dockit/grid"
)

type Loader interface {
	Name() string
	Detect(string) (bool, error)
	Open(string) (grid.File, error)

	New() (grid.File, error)
	IsSupportedExt(string) bool
}

type Merger interface {
	Merge(grid.File) error
}
