package builtins

import (
	"math"

	_ "github.com/midbel/dockit/grid/text"
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
	return value.FindIndex(args[1:], args[0])
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
	arr, ok := args[0].(value.ArrayValue)
	if !ok {
		return value.ErrValue
	}
	var (
		row = asFloat(args[1])
		col = 1.0
	)
	if len(args) >= 3 {
		col = asFloat(args[2])
	}
	return arr.At(int(math.Floor(row)-1), int(math.Floor(col)-1))
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
