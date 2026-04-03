package builtins

import (
	"math"

	"github.com/midbel/dockit/grid/calc"
	"github.com/midbel/dockit/grid/criteria"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var signBuiltin = Builtin{
	Name:     "sign",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Sign,
}

func Sign(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	s := calc.Sign(asFloat(args[0]))
	return value.Float(s)
}

var isOddBuiltin = Builtin{
	Name:     "isodd",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: IsOdd,
}

func IsOdd(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
	return value.Boolean(calc.Odd(v))
}

var isEvenBuiltin = Builtin{
	Name:     "iseven",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: IsEven,
}

func IsEven(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
	return value.Boolean(calc.Even(v))
}

var minBuiltin = Builtin{
	Name:     "min",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Min,
}

func Min(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Min(arr)
	)
	return value.Float(ret)
}

var maxBuiltin = Builtin{
	Name:     "max",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Max,
}

func Max(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Max(arr)
	)
	return value.Float(ret)
}

var sumBuiltin = Builtin{
	Name:     "sum",
	Desc:     "Returns the sum of all values",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Sum,
}

func Sum(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Sum(arr)
	)
	return value.Float(ret)
}

var sumifBuiltin = Builtin{
	Name:     "sumif",
	Desc:     "Sums values that match a condition",
	Category: "miscel",
	Params: []Param{
		Array("value", "", value.TypeAny),
		Scalar("predicate", "", value.TypeText),
	},
	Func: SumIf,
}

func SumIf(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f, err := criteria.New(asString(args[1]))
	if err != nil {
		return value.ErrValue
	}
	total := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		if ok := f.Keep(v); ok {
			acc += asFloat(v)
		}
		return acc
	})
	return value.Float(total)
}

var avgBuiltin = Builtin{
	Name:     "average",
	Alias:    slx.Make("avg"),
	Desc:     "Returns the average of the given values",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Avg,
}

func Avg(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Avg(arr)
	)
	return value.Float(ret)
}

var avgifBuiltin = Builtin{
	Name:     "averageif",
	Desc:     "",
	Category: "miscel",
	Params: []Param{
		Array("value", "", value.TypeAny),
		Scalar("predicate", "", value.TypeText),
	},
	Func: AvgIf,
}

func AvgIf(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f, err := criteria.New(asString(args[1]))
	if err != nil {
		return value.ErrValue
	}
	vs := value.Collect[float64](args, func(v value.Value) (float64, bool) {
		return asFloat(v), f.Keep(v)
	})
	ret := calc.Avg(vs)
	return value.Float(ret)
}

var stdevBuiltin = Builtin{
	Name:     "stdev",
	Desc:     "",
	Category: "math",
	Func:     Stdev,
}

func Stdev(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Stdev(arr)
	)
	return value.Float(ret)
}

var varianceBuiltin = Builtin{
	Name:     "var",
	Alias:    slx.Make("variance"),
	Desc:     "",
	Category: "math",
	Func:     Variance,
}

func Variance(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Var(arr)
	)
	return value.Float(ret)
}

var modeBuiltin = Builtin{
	Name:     "mode",
	Desc:     "",
	Category: "math",
	Func:     Mode,
}

func Mode(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Mode(arr)
	)
	return value.Float(ret)
}

var medianBuiltin = Builtin{
	Name:     "median",
	Desc:     "",
	Category: "math",
	Func:     Median,
}

func Median(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Median(arr)
	)
	return value.Float(ret)
}

var countBuiltin = Builtin{
	Name:     "count",
	Desc:     "Counts numeric values",
	Category: "miscel",
	Params: []Param{
		Var(ScalarArray("value", "", value.TypeAny)),
	},
	Func: Count,
}

func Count(args []value.Value) value.Value {
	count := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		return acc + 1
	})
	return value.Float(count)
}

var countifBuiltin = Builtin{
	Name:     "countif",
	Desc:     "Counts values that match a condition",
	Category: "miscel",
	Params: []Param{
		Array("value", "", value.TypeAny),
		Scalar("predicate", "", value.TypeText),
	},
	Func: CountIf,
}

func CountIf(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f, err := criteria.New(asString(args[1]))
	if err != nil {
		return value.ErrValue
	}
	count := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		if ok := f.Keep(v); ok {
			acc += 1
		}
		return acc
	})
	return value.Float(count)
}

var countaBuiltin = Builtin{
	Name:     "counta",
	Desc:     "Counts non-empty values",
	Category: "miscel",
	Params: []Param{
		Var(ScalarArray("value", "", value.TypeAny)),
	},
	Func: Count,
}

func Counta(args []value.Value) value.Value {
	count := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		if value.IsBlank(v) || asString(v) == "" {
			return acc
		}
		return acc + 1
	})
	return value.Float(count)
}

var roundBuiltin = Builtin{
	Name:     "round",
	Desc:     "Rounds a number to the nearest integer",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Round,
}

func Round(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Round(f)
	return value.Float(ret)
}

var floorBuiltin = Builtin{
	Name:     "rounddown",
	Alias:    slx.Make("floor"),
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Floor,
}

func Floor(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Floor(f)
	return value.Float(ret)
}

var ceilBuiltin = Builtin{
	Name:     "roundup",
	Alias:    slx.Make("ceil"),
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Ceil,
}

func Ceil(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Ceil(f)
	return value.Float(ret)
}

var sqrtBuiltin = Builtin{
	Name:     "sqrt",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Sqrt,
}

func Sqrt(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	if f < 0 {
		return value.ErrValue
	}
	ret := math.Sqrt(f)
	return value.Float(ret)
}

var absBuiltin = Builtin{
	Name:     "abs",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Abs,
}

func Abs(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Abs(f)
	return value.Float(ret)
}

var modBuiltin = Builtin{
	Name:     "mod",
	Desc:     "Returns the remainder after division",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
		Scalar("number", "", value.TypeNumber),
	},
	Func: Mod,
}

func Mod(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		d = asFloat(args[1])
	)
	if d == 0 {
		return value.ErrDiv0
	}
	ret := math.Mod(f, d)
	return value.Float(ret)
}

var powBuiltin = Builtin{
	Name:     "power",
	Alias:    slx.Make("pow"),
	Desc:     "Raises a number to a given power",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
		Scalar("number", "", value.TypeNumber),
	},
	Func: Pow,
}

func Pow(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		e = asFloat(args[1])
	)
	ret := math.Pow(f, e)
	return value.Float(ret)
}

var intBuiltin = Builtin{
	Name:     "int",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Int,
}

func Int(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	return value.Float(int(f))
}

var randBuiltin = Builtin{
	Name:     "rand",
	Alias:    slx.Make("random"),
	Category: "math",
	Params:   []Param{},
	Func:     Rand,
}

func Rand(args []value.Value) value.Value {
	r := calc.Rand()
	return value.Float(r)
}

var sinBuiltin = Builtin{
	Name:     "sin",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Sin,
}

func Sin(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Sin(f)
	)
	return value.Float(r)
}

var cosBuiltin = Builtin{
	Name:     "cos",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Cos,
}

func Cos(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Cos(f)
	)
	return value.Float(r)
}

var tanBuiltin = Builtin{
	Name:     "tan",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Tan,
}

func Tan(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Tan(f)
	)
	return value.Float(r)
}

var asinBuiltin = Builtin{
	Name:     "asin",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Asin,
}

func Asin(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Asin(f)
	)
	return value.Float(r)
}

var acosBuiltin = Builtin{
	Name:     "acos",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Acos,
}

func Acos(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Acos(f)
	)
	return value.Float(r)
}

var atan2Builtin = Builtin{
	Name:     "atan2",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("xnum", "", value.TypeNumber)),
		Var(ScalarArray("ynum", "", value.TypeNumber)),
	},
	Func: Atan2,
}

func Atan2(args []value.Value) value.Value {
	var (
		vx  = asFloat(args[0])
		vy  = asFloat(args[1])
		ret = math.Atan2(vx, vy)
	)
	return value.Float(ret)
}

var degBuiltin = Builtin{
	Name:     "degress",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Deg,
}

func Deg(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	r := calc.Deg(asFloat(args[0]))
	return value.Float(r)
}

var radBuiltin = Builtin{
	Name:     "radians",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Scalar("number", "", value.TypeNumber),
	},
	Func: Rad,
}

func Rad(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	r := calc.Rad(asFloat(args[0]))
	return value.Float(r)
}

var piBuiltin = Builtin{
	Name:     "pi",
	Desc:     "",
	Category: "math",
	Func:     Pi,
}

func Pi(args []value.Value) value.Value {
	pi := calc.Pi()
	return value.Float(pi)
}

var log10Builtin = Builtin{
	Name:     "log10",
	Desc:     "Returns the base-10 logarithm of a number",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Log10,
}

func Log10(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Log10(f)
	)
	return value.Float(r)
}

var lnBuiltin = Builtin{
	Name:     "ln",
	Desc:     "Returns the natural logarithm of a number",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Ln,
}

func Ln(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Log(f)
	)
	return value.Float(r)
}

var expBuiltin = Builtin{
	Name:     "exp",
	Desc:     "",
	Category: "math",
	Params: []Param{
		Var(ScalarArray("number", "", value.TypeNumber)),
	},
	Func: Exp,
}

func Exp(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Exp(f)
	)
	return value.Float(r)
}

var eBuiltin = Builtin{
	Name:     "e",
	Desc:     "",
	Category: "math",
	Func:     E,
}

func E(args []value.Value) value.Value {
	e := calc.E()
	return value.Float(e)
}

var numberBuiltins = []Builtin{
	signBuiltin,
	isOddBuiltin,
	isEvenBuiltin,
	minBuiltin,
	maxBuiltin,
	sumBuiltin,
	sumifBuiltin,
	avgBuiltin,
	avgifBuiltin,
	stdevBuiltin,
	varianceBuiltin,
	modeBuiltin,
	medianBuiltin,
	countBuiltin,
	countifBuiltin,
	countaBuiltin,
	roundBuiltin,
	floorBuiltin,
	ceilBuiltin,
	sqrtBuiltin,
	absBuiltin,
	modBuiltin,
	powBuiltin,
	intBuiltin,
	randBuiltin,
	sinBuiltin,
	cosBuiltin,
	tanBuiltin,
	asinBuiltin,
	acosBuiltin,
	atan2Builtin,
	degBuiltin,
	radBuiltin,
	piBuiltin,
	log10Builtin,
	lnBuiltin,
	expBuiltin,
	eBuiltin,
}
