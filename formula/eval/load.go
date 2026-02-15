package eval

import (
	"fmt"

	"github.com/midbel/dockit/grid"
)

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
