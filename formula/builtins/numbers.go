package builtins

import (
	"time"

	gbs "github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/internal/slx"
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
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		step  = asFloat(args[1])
		count = asFloat(args[2])
	)
	switch a := args[0].(type) {
	case value.Date:
		return generateDateSeq(time.Time(a), int64(step), int64(count))
	case value.Float:
		return generateNumberSeq(float64(a), int64(step), int64(count))
	default:
		return value.ErrValue
	}
}

func generateDateSeq(start time.Time, step, count int64) value.Value {
	arr := make([][]value.Value, count)
	for i := int64(0); i < count; i++ {
		var (
			t = start.Add(time.Duration(step*24) * time.Hour)
			x value.Value
		)
		x = value.Date(t)
		arr[i] = slx.One(x)
	}
	return value.NewArray(arr)
}

func generateNumberSeq(start float64, step, count int64) value.Value {
	arr := make([][]value.Value, count)
	for i := int64(0); i < count; i++ {
		var (
			t = start + float64(i*step)
			x value.Value
		)
		x = value.Float(t)
		arr[i] = slx.One(x)
	}
	return value.NewArray(arr)
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
