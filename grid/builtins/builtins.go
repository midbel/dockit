package builtins

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var (
	ErrArity = errors.New("invalid number of arguments")
	ErrType  = errors.New("invalid type")
)

type ValueIterator interface {
	Values() iter.Seq[value.ScalarValue]
}

func Each(args []value.Value, fn func(value.Value) error) error {
	for _, a := range args {
		if value.IsScalar(a) {
			if err := fn(a); err != nil {
				return err
			}
		} else if value.IsArray(a) {
			it, ok := a.(ValueIterator)
			if !ok {
				return fmt.Errorf("array does not implement value iterator")
			}
			var dat []value.Value
			for v := range it.Values() {
				dat = append(dat, v)
			}
			if err := Each(dat, fn); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported value type")
		}
	}
	return nil
}

var Registry = map[string]Builtin{
	"min": {
		Name:     "min",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Min,
	},
	"max": {
		Name:     "max",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Max,
	},
	"sum": {
		Name:     "sum",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Sum,
	},
	"average": {
		Name:     "average",
		Alias:    slx.Make("avg"),
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Avg,
	},
	"round": {
		Name:     "round",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Round,
	},
	"rounddown": {
		Name:     "rounddown",
		Alias:    slx.Make("floor"),
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Floor,
	},
	"roundup": {
		Name:     "rounddown",
		Alias:    slx.Make("ceil"),
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Ceil,
	},
	"sqrt": {
		Name:     "sqrt",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Sqrt,
	},
	"abs": {
		Name:     "abs",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Abs,
	},
	"mod": {
		Name:     "mod",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
			Scalar("number", "", value.TypeNumber),
		},
		Func: Mod,
	},
	"power": {
		Name:     "power",
		Alias:    slx.Make("pow"),
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
			Scalar("number", "", value.TypeNumber),
		},
		Func: Pow,
	},
	"int": {
		Name:     "int",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Int,
	},
	"rand": {
		Name:     "rand",
		Alias:    slx.Make("random"),
		Category: "math",
		Params:   []Param{},
		Func:     Rand,
	},
	// "sumif":       nil,
	// "countif":     nil,
	"count": {
		Name:     "count",
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Array("value", "", value.TypeAny),
		},
		Func: Count,
	},
	"type": {
		Name:     "type",
		Alias:    slx.Make("typeof"),
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
		},
		Func: TypeOf,
	},
	"now": {
		Name:     "now",
		Desc:     "",
		Category: "time",
		Params:   []Param{},
		Func:     Now,
	},
	"today": {
		Name:     "today",
		Desc:     "",
		Category: "time",
		Params:   []Param{},
		Func:     Today,
	},
	"date": {
		Name:     "date",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("year", "", value.TypeNumber),
			Scalar("month", "", value.TypeNumber),
			Scalar("day", "", value.TypeNumber),
		},
		Func: Date,
	},
	"year": {
		Name:     "year",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Year,
	},
	"month": {
		Name:     "month",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Month,
	},
	"day": {
		Name:     "day",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Day,
	},
	"yearday": {
		Name:     "yearday",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: YearDay,
	},
	"hour": {
		Name:     "hour",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Hour,
	},
	"minute": {
		Name:     "minute",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Minute,
	},
	"second": {
		Name:     "second",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Second,
	},
	"weekday": {
		Name:     "weekday",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("date", "", value.TypeDate),
		},
		Func: Weekday,
	},
	"isnumber": {
		Name:     "isnumber",
		Desc:     "",
		Category: "util",
		Params: []Param{
			ScalarArray("value", "", value.TypeAny),
		},
		Func: IsNumber,
	},
	"istext": {
		Name:     "istext",
		Desc:     "",
		Category: "util",
		Params: []Param{
			ScalarArray("value", "", value.TypeAny),
		},
		Func: IsText,
	},
	// "isblank":     nil,
	// "iserror":     nil,
	// "isna":        nil,
	"concatenate": {
		Name:     "concatenate",
		Desc:     "",
		Category: "text",
		Alias:    slx.Make("concat"),
		Params: []Param{
			Var(Scalar("str", "", value.TypeText)),
		},
		Func: Concat,
	},
	"left": {
		Name:     "left",
		Desc:     "",
		Category: "text",
		Params:   []Param{},
		Func:     Left,
	},
	"right": {
		Name:     "right",
		Desc:     "",
		Category: "text",
		Params:   []Param{},
		Func:     Right,
	},
	"mid": {
		Name:     "mid",
		Desc:     "",
		Category: "text",
		Params:   []Param{},
		Func:     Mid,
	},
	"len": {
		Name:     "len",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Len,
	},
	"upper": {
		Name:     "upper",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Upper,
	},
	"lower": {
		Name:     "lower",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Lower,
	},
	"substr": {
		Name:     "substr",
		Desc:     "",
		Category: "text",
		Params:   []Param{},
		Func:     Substr,
	},
	"replace": {
		Name:     "replace",
		Desc:     "",
		Category: "text",
		Params:   []Param{},
		Func:     Replace,
	},
	// "trim":        nil,
	// "split":       nil,
	// "join":        nil,
	// "proper":      nil,
	// "search":      nil,
	// "find":        nil,
	// "substitute":  nil,
	// "text":        nil,
	// "value":       nil,
	// "textjoin":    nil,
	"if": {
		Name:     "if",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     If,
	},
	// "iferror":     nil,
	// "ifs":         nil,
	// "ifna":        nil,
	"and": {
		Name:     "and",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     And,
	},
	"or": {
		Name:     "or",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     Or,
	},
	"xor": {
		Name:     "xor",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     Xor,
	},
	"not": {
		Name:     "not",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     Not,
	},
	// "index":       nil,
	// "match":       nil,
	// "vlookup":     nil,
	// "hlookup":     nil,
	// "xlookup":     nil,
	// "offset":      nil,
	// "choose":      nil,
	"lock": {
		Name:     "lock",
		Desc:     "",
		Category: "miscel",
		Params:   []Param{},
		Func:     Lock,
	},
	"unlock": {
		Name:     "unlock",
		Desc:     "",
		Category: "miscel",
		Params:   []Param{},
		Func:     Unlock,
	},
}

func Lookup(ident string) (BuiltinFunc, error) {
	fn, ok := Registry[strings.ToLower(ident)]
	if ok {
		return fn.Make(), nil
	}
	vs := slices.Collect(maps.Values(Registry))
	ix := slices.IndexFunc(vs, func(b Builtin) bool {
		return slices.Contains(b.Alias, ident)
	})
	if ix < 0 {
		return nil, fmt.Errorf("%s undefined builtin", ident)
	}
	return vs[ix].Make(), nil
}

func List() []Builtin {
	vs := maps.Values(Registry)
	return slices.Collect(vs)
}

type BuiltinFunc func([]value.Value) value.Value

type Builtin struct {
	Name     string
	Desc     string
	Category string
	Alias    []string
	Params   []Param
	Func     BuiltinFunc

	Dialect parse.Dialect
}

func (b Builtin) OxmlSupported() bool {
	if b.Dialect == 0 {
		return true
	}
	return b.Dialect&parse.OxmlDialect != 0
}

func (b Builtin) OdsSupported() bool {
	if b.Dialect == 0 {
		return true
	}
	return b.Dialect&parse.OdsDialect != 0
}

func (b Builtin) OxmlOnly() bool {
	return b.Dialect&parse.OxmlDialect == parse.OxmlDialect
}

func (b Builtin) OdsOnly() bool {
	return b.Dialect&parse.OdsDialect == parse.OdsDialect
}

func (b Builtin) Make() BuiltinFunc {
	return Make(b.Params, b.Func)
}

type Param struct {
	Name     string
	Desc     string
	Type     string
	Mode     value.ValueKind
	Optional bool
	Variadic bool
}

func (p Param) Valid(val value.Value) bool {
	kd := val.Kind()
	return kd&p.Mode != 0
}

func (p Param) Convert(val value.Value) value.Value {
	if !p.Valid(val) {
		return value.ErrValue
	}
	if value.IsArray(val) {
		arr, ok := val.(value.Array)
		if !ok {
			return value.ErrValue
		}
		apply := func(v value.ScalarValue) value.ScalarValue {
			ret := p.Value(v)
			if ret.Kind() != value.KindScalar {
				return value.ErrValue
			}
			return ret.(value.ScalarValue)
		}

		other := arr.Clone()
		other.Apply(apply)
		return other
	}
	return p.Value(val)
}

func (p Param) Value(val value.Value) value.Value {
	var (
		ret value.Value
		err error
	)
	switch p.Type {
	case value.TypeNumber:
		ret, err = value.CastToFloat(val)
	case value.TypeText:
		ret, err = value.CastToText(val)
	case value.TypeBool:
		ok := value.True(val)
		ret = value.Boolean(ok)
	case value.TypeDate:
		ret, err = value.CastToDate(val)
	case value.TypeAny:
		return val
	default:
		return value.ErrValue
	}
	if err != nil {
		ret = value.ErrValue
	}
	return ret
}

func Make(params []Param, do BuiltinFunc) BuiltinFunc {
	fn := func(args []value.Value) value.Value {
		var (
			newArgs []value.Value
			pix     int
		)
		for aix := 0; aix < len(args); aix++ {
			if pix >= len(params) {
				return value.ErrValue
			}
			var (
				p  = params[pix]
				as []value.Value
			)
			if p.Variadic && pix == len(params)-1 {
				as = args[aix:]
				aix += len(as)
			} else {
				as = args[aix : aix+1]
			}
			for i := range as {
				ret := p.Convert(as[i])
				if value.IsError(ret) {
					return ret
				}
				newArgs = append(newArgs, ret)
			}
			pix++
		}
		for _, p := range params[pix:] {
			if !p.Optional && !p.Variadic && len(newArgs) < len(params) {
				return value.ErrValue
			}
		}
		return do(newArgs)
	}
	return fn
}

func Scalar(name, desc string, k string) Param {
	return Param{
		Name: name,
		Desc: desc,
		Type: k,
		Mode: value.KindScalar,
	}
}

func Array(name, desc string, k string) Param {
	return Param{
		Name: name,
		Desc: desc,
		Type: k,
		Mode: value.KindArray,
	}
}

func ScalarArray(name, desc string, k string) Param {
	return Param{
		Name: name,
		Desc: desc,
		Type: k,
		Mode: value.KindScalar | value.KindArray,
	}
}

func Opt(p Param) Param {
	p.Optional = true
	return p
}

func Var(p Param) Param {
	p.Variadic = true
	return p
}
