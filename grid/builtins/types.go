package builtins

import (
	"github.com/midbel/dockit/value"
)

func TypeOf(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, ErrArity
	}
	return value.Text(args[0].Type()), nil
}
