package builtins

import (
	gbs "github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/value"
)

var seqBuiltin = gbs.Builtin{
	Name: "seq",
	Desc: "",
	Params: []gbs.Param{
		gbs.Scalar("start", "", value.TypeAny),
		gbs.Scalar("step", "", value.TypeNumber),
		gbs.Scalar("count", "", value.TypeNumber),
	},
	Category: "numbers",
	Func:     Seq,
}

func Seq(args []value.Value) value.Value {
	return nil
}

var rangeBuiltin = gbs.Builtin{
	Name: "range",
	Desc: "",
	Params: []gbs.Param{
		gbs.Scalar("start", "", value.TypeAny),
		gbs.Scalar("end", "", value.TypeAny),
		gbs.Scalar("step", "", value.TypeNumber),
	},
	Category: "numbers",
	Func:     Range,
}

func Range(args []value.Value) value.Value {
	return nil
}

var numberBuiltins = []gbs.Builtin{
	seqBuiltin,
	rangeBuiltin,
}
