package criteria

import (
	"github.com/midbel/dockit/value"
)

type Filter interface {
	Keep(value.Value) bool
}

func Compile(str string) (Filter, error) {
	return nil, nil
}
