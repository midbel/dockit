package builtins

import (
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var typeofBuiltin = Builtin{
	Name:     "type",
	Alias:    slx.Make("typeof"),
	Desc:     "",
	Category: "miscel",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
	},
	Func: TypeOf,
}

func TypeOf(args []value.Value) value.Value {
	return value.Text(args[0].Type())
}

var isNumberBuiltin = Builtin{
	Name:     "isnumber",
	Desc:     "",
	Category: "util",
	Params: []Param{
		ScalarArray("value", "", value.TypeAny),
	},
	Func: IsNumber,
}

func IsNumber(args []value.Value) value.Value {
	ok := value.IsNumber(args[0])
	return value.Boolean(ok)
}

var isTextBuiltin = Builtin{
	Name:     "istext",
	Desc:     "",
	Category: "util",
	Params: []Param{
		ScalarArray("value", "", value.TypeAny),
	},
	Func: IsText,
}

func IsText(args []value.Value) value.Value {
	ok := value.IsText(args[0])
	return value.Boolean(ok)
}

var isBlankBuiltin = Builtin{
	Name:     "isblank",
	Desc:     "",
	Category: "type",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
	},
	Func: IsBlank,
}

func IsBlank(args []value.Value) value.Value {
	ok := value.IsBlank(args[0])
	return value.Boolean(ok)
}

var isErrorBuiltin = Builtin{
	Name:     "iserror",
	Desc:     "",
	Category: "type",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
	},
	Func: IsError,
}

func IsError(args []value.Value) value.Value {
	ok := value.IsError(args[0])
	return value.Boolean(ok)
}

var isNaBuiltin = Builtin{
	Name:     "isna",
	Desc:     "",
	Category: "type",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
	},
	Func: IsNA,
}

func IsNA(args []value.Value) value.Value {
	ok := value.IsError(args[0]) && args[0] == value.ErrNA
	return value.Boolean(ok)
}

var naBuiltin = Builtin{
	Name:     "na",
	Desc:     "",
	Category: "errors",
	Func:     Na,
}

func Na(args []value.Value) value.Value {
	return value.ErrNA
}

var errBuiltin = Builtin{
	Name:     "err",
	Desc:     "",
	Category: "errors",
	Params: []Param{
		Scalar("str", "", value.TypeText),
	},
	Func: Err,
}

func Err(args []value.Value) value.Value {
	switch asString(args[0]) {
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

var typeBuiltins = []Builtin{
	typeofBuiltin,
	isBlankBuiltin,
	isErrorBuiltin,
	isTextBuiltin,
	isNumberBuiltin,
	isNaBuiltin,
	naBuiltin,
	errBuiltin,
}
