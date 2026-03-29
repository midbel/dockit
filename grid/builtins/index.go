package builtins

import (
	"github.com/midbel/dockit/value"
)

var matchBuiltin = Builtin{
	Name:     "match",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Scalar("value", "", value.TypeAny),
		Array("array", "", value.TypeAny),
	},
	Func: Match,
}

func Match(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	return value.FindIndex(args[1], args[0])
}

var indexBuiltin = Builtin{
	Name:     "index",
	Desc:     "",
	Category: "conditional",
	Params: []Param{
		Array("array", "", value.TypeAny),
		Scalar("row", "", value.TypeAny),
		Opt(Scalar("col", "", value.TypeAny)),
	},
	Func: Index,
}

func Index(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	return nil
}

var vlookupBuiltin = Builtin{
	Name:     "vlookup",
	Desc:     "",
	Category: "conditional",
	Params:   []Param{},
	Func:     VLookup,
}

func VLookup(args []value.Value) value.Value {
	return nil
}

func HLookup(args []value.Value) value.Value {
	return nil
}

func XLookup(args []value.Value) value.Value {
	return nil
}

var indexBuiltins = []Builtin{
	matchBuiltin,
	indexBuiltin,
	vlookupBuiltin,
}
