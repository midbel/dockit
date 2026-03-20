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

func Err(args []value.Value) value.Value {
	str, _ := value.CastToText(args[0])
	switch str {
	case "Null":
		return value.ErrNull
	case "Div0":
		return value.ErrDiv0
	case "Value":
		return value.ErrValue
	case "Ref":
		return value.ErrRef
	case "Name":
		return value.ErrName
	case "Num":
		return value.ErrNum
	default:
		return value.ErrNA
	}
}
