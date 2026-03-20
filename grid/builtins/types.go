package builtins

import (
	"github.com/midbel/dockit/value"
)

func TypeOf(args []value.Value) value.Value {
	return value.Text(args[0].Type())
}

func IsBlank(args []value.Value) value.Value {
	ok := value.IsBlank(args[0])
	return value.Boolean(ok)
}

func IsError(args []value.Value) value.Value {
	ok := value.IsError(args[0])
	return value.Boolean(ok)
}

func IsNA(args []value.Value) value.Value {
	ok := value.IsError(args[0]) && args[0] == value.ErrNA
	return value.Boolean(ok)
}

func Na(args []value.Value) value.Value {
	return value.ErrNA
}
