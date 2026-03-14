package builtins

import (
	"github.com/midbel/dockit/value"
)

func TypeOf(args []value.Value) value.Value {
	return value.Text(args[0].Type())
}
