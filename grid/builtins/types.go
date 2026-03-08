package builtins

import (
	"github.com/midbel/dockit/value"
)

func TypeOf(args []value.Value) (value.Value, error) {
	return value.Text(args[0].Type()), nil
}
