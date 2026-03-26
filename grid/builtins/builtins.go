package builtins

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
	"github.com/midbel/textwrap"
)

var (
	ErrArity = errors.New("invalid number of arguments")
	ErrType  = errors.New("invalid type")
)

var Registry = map[string]Builtin{
	"e": {
		Name:     "e",
		Desc:     "",
		Category: "math",
		Func:     E,
	},
	"exp": {
		Name:     "exp",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Exp,
	},
	"ln": {
		Name:     "ln",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Ln,
	},
	"log10": {
		Name:     "log10",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Log10,
	},
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
	"isodd": {
		Name:     "isodd",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: IsOdd,
	},
	"iseven": {
		Name:     "iseven",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: IsEven,
	},
	"sign": {
		Name:     "sign",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Sign,
	},
	"sin": {
		Name:     "sin",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Sin,
	},
	"cos": {
		Name:     "cos",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Cos,
	},
	"tan": {
		Name:     "tan",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Tan,
	},
	"asin": {
		Name:     "asin",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Asin,
	},
	"acos": {
		Name:     "acos",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("number", "", value.TypeNumber)),
		},
		Func: Acos,
	},
	"atan2": {
		Name:     "atan2",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Var(ScalarArray("xnum", "", value.TypeNumber)),
			Var(ScalarArray("ynum", "", value.TypeNumber)),
		},
		Func: Atan2,
	},
	"pi": {
		Name:     "pi",
		Desc:     "",
		Category: "math",
		Func:     Pi,
	},
	"degrees": {
		Name:     "degress",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Deg,
	},
	"radians": {
		Name:     "radians",
		Desc:     "",
		Category: "math",
		Params: []Param{
			Scalar("number", "", value.TypeNumber),
		},
		Func: Rad,
	},
	"sum": {
		Name:     "sum",
		Desc:     "Returns the sum of all values",
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
	"stddev": {
		Name:     "stddev",
		Desc:     "",
		Category: "math",
		Func:     Stdev,
	},
	"mode": {
		Name:     "mode",
		Desc:     "",
		Category: "math",
		Func:     Mode,
	},
	"median": {
		Name:     "median",
		Desc:     "",
		Category: "math",
		Func:     Median,
	},
	"var": {
		Name:     "var",
		Alias:    slx.Make("variance"),
		Desc:     "",
		Category: "math",
		Func:     Variance,
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
		Name:     "roundup",
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
	"counta": {
		Name:     "counta",
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Var(ScalarArray("value", "", value.TypeAny)),
		},
		Func: Count,
	},
	"sumif": {
		Name:     "sumif",
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Array("value", "", value.TypeAny),
			Scalar("predicate", "", value.TypeText),
		},
		Func: SumIf,
	},
	"countif": {
		Name:     "countif",
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Array("value", "", value.TypeAny),
			Scalar("predicate", "", value.TypeText),
		},
		Func: CountIf,
	},
	"averageif": {
		Name:     "averageif",
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Array("value", "", value.TypeAny),
			Scalar("predicate", "", value.TypeText),
		},
		Func: AvgIf,
	},
	"count": {
		Name:     "count",
		Desc:     "",
		Category: "miscel",
		Params: []Param{
			Var(ScalarArray("value", "", value.TypeAny)),
		},
		Func: Count,
	},
	"na": {
		Name:     "na",
		Desc:     "",
		Category: "errors",
		Func:     Na,
	},
	"err": {
		Name:     "err",
		Desc:     "",
		Category: "errors",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Err,
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
	"datedif": {
		Name:     "datedif",
		Desc:     "",
		Category: "time",
		Params: []Param{
			Scalar("fromDate", "", value.TypeDate),
			Scalar("toDate", "", value.TypeDate),
			Scalar("diffUnit", "", value.TypeText),
		},
		Func: DateDiff,
	},
	"edate": {
		Name:     "edate",
		Desc:     "",
		Category: "time",
		Params:   []Param{},
		Func:     Edate,
	},
	"eomonth": {
		Name:     "eomonth",
		Desc:     "",
		Category: "time",
		Params:   []Param{},
		Func:     EoMonth,
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
	"isblank": {
		Name:     "isblank",
		Desc:     "",
		Category: "type",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
		},
		Func: IsBlank,
	},
	"iserror": {
		Name:     "iserror",
		Desc:     "",
		Category: "type",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
		},
		Func: IsError,
	},
	"isna": {
		Name:     "isna",
		Desc:     "",
		Category: "type",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
		},
		Func: IsNA,
	},
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
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Opt(Scalar("chars", "", value.TypeNumber)),
		},
		Func: Left,
	},
	"right": {
		Name:     "right",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Opt(Scalar("chars", "", value.TypeNumber)),
		},
		Func: Right,
	},
	"mid": {
		Name:     "mid",
		Desc:     "",
		Alias:    []string{"substr"},
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("index", "", value.TypeNumber),
			Scalar("chars", "", value.TypeNumber),
		},
		Func: Mid,
	},
	"len": {
		Name:     "len",
		Desc:     "",
		Category: "text",
		Alias:    []string{"length"},
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
	"trim": {
		Name:     "trim",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Trim,
	},
	"clean": {
		Name:     "clean",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Clean,
	},
	"search": {
		Name:     "search",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("find", "", value.TypeText),
			Opt(Scalar("offset", "", value.TypeNumber)),
		},
		Func: Search,
	},
	"find": {
		Name:     "find",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("find", "", value.TypeText),
			Opt(Scalar("offset", "", value.TypeNumber)),
		},
		Func: Find,
	},
	"proper": {
		Name:     "proper",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Proper,
	},
	"rept": {
		Name:     "rept",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("num", "", value.TypeNumber),
		},
		Func: Rept,
	},
	"substitute": {
		Name:     "substitue",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("old", "", value.TypeText),
			Scalar("new", "", value.TypeText),
			Opt(Scalar("num", "", value.TypeNumber)),
		},
		Func: Substitute,
	},
	"replace": {
		Name:     "replace",
		Desc:     "Replace part of text at given position with new text",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("pos", "", value.TypeNumber),
			Scalar("num", "", value.TypeNumber),
			Scalar("new", "", value.TypeText),
		},
		Func: Replace,
	},
	"text": {
		Name:     "text",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
			Scalar("pattern", "", value.TypeText),
		},
		Func: Text,
	},
	"value": {
		Name:     "value",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("str", "", value.TypeText),
		},
		Func: Text,
	},
	"textjoin": {
		Name:     "textjoin",
		Desc:     "",
		Category: "text",
		Params: []Param{
			Scalar("delimiter", "", value.TypeText),
			Scalar("ignore", "", value.TypeBool),
			Var(Scalar("str", "", value.TypeText)),
		},
		Func: Text,
	},
	"choose": {
		Name:     "choose",
		Desc:     "Returns the value at the given 1-based index. If the index is out of range, returns ErrNA",
		Category: "conditional",
		Params: []Param{
			Scalar("index", "", value.TypeNumber),
			Deferrable(Var(Scalar("value", "", value.TypeAny))),
		},
		Func: Choose,
	},
	"switch": {
		Name:     "switch",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			Scalar("var", "", value.TypeNumber),
			Var(Scalar("value", "", value.TypeAny)),
			Opt(Scalar("default", "", value.TypeAny)),
		},
		Func: Switch,
	},
	"match": {
		Name:     "match",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     Match,
	},
	"index": {
		Name:     "index",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     Index,
	},
	"vlookup": {
		Name:     "vlookup",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     VLookup,
	},
	"ifs": {
		Name:     "ifs",
		Desc:     "",
		Category: "conditional",
		Params:   []Param{},
		Func:     Ifs,
	},
	"if": {
		Name:     "if",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
			Deferrable(Scalar("csq", "", value.TypeAny)),
			Deferrable(Scalar("alt", "", value.TypeAny)),
		},
		Func: If,
	},
	"iferror": {
		Name:     "iferror",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
			Deferrable(ScalarArray("replace", "", value.TypeAny)),
		},
		Func: IfError,
	},
	"ifna": {
		Name:     "ifna",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			Scalar("value", "", value.TypeAny),
			Deferrable(ScalarArray("replace", "", value.TypeAny)),
		},
		Func: IfNA,
	},
	"and": {
		Name:     "and",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			ScalarArray("value1", "", value.TypeAny),
			ScalarArray("value2", "", value.TypeAny),
		},
		Func: And,
	},
	"or": {
		Name:     "or",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			ScalarArray("value1", "", value.TypeAny),
			ScalarArray("value2", "", value.TypeAny),
		},
		Func: Or,
	},
	"xor": {
		Name:     "xor",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			Var(ScalarArray("value", "", value.TypeAny)),
		},
		Func: Xor,
	},
	"not": {
		Name:     "not",
		Desc:     "",
		Category: "conditional",
		Params: []Param{
			ScalarArray("value", "", value.TypeAny),
		},
		Func: Not,
	},
}

func Help(w io.Writer, ident string) error {
	b, err := Get(ident)
	if err != nil {
		return err
	}
	ws := bufio.NewWriter(w)
	defer ws.Flush()
	io.WriteString(ws, strings.ToUpper(b.Name))
	io.WriteString(ws, "(")
	for i, p := range b.Params {
		if i > 0 {
			io.WriteString(ws, ", ")
		}
		if p.Optional {
			io.WriteString(ws, "[")
		}
		io.WriteString(ws, strings.ToUpper(p.Name))
		io.WriteString(ws, ": ")
		io.WriteString(ws, p.Type)
		if p.Variadic {
			io.WriteString(ws, "...")
		}
		if p.Optional {
			io.WriteString(ws, "]")
		}
	}
	io.WriteString(ws, ")")

	io.WriteString(ws, "\t[")
	io.WriteString(ws, b.Category)
	io.WriteString(ws, "]")
	io.WriteString(ws, "\n")
	io.WriteString(ws, "\n")

	io.WriteString(ws, textwrap.WrapN(b.Desc, 70))
	io.WriteString(ws, ".\n")
	io.WriteString(ws, "\n")

	io.WriteString(ws, "PARAMETERS:")
	io.WriteString(ws, "\n")
	for _, p := range b.Params {
		io.WriteString(ws, " "+strings.ToUpper(p.Name))
		io.WriteString(ws, ": ")
		io.WriteString(ws, p.Desc)
		io.WriteString(ws, "\n")
	}
	io.WriteString(ws, "\n")
	io.WriteString(ws, "\n")
	io.WriteString(ws, "SUPPORT:")
	io.WriteString(ws, "\n")
	io.WriteString(ws, " OXML: ")
	io.WriteString(ws, cli.MarkBool(b.OxmlSupported()))
	io.WriteString(ws, "\n")
	io.WriteString(ws, " ODS : ")
	io.WriteString(ws, cli.MarkBool(b.OdsSupported()))
	io.WriteString(ws, "\n")
	io.WriteString(ws, "\n")
	return nil
}

func Get(ident string) (Builtin, error) {
	fn, ok := Registry[strings.ToLower(ident)]
	if ok {
		return fn, nil
	}
	vs := List()
	ix := slices.IndexFunc(vs, func(b Builtin) bool {
		return slices.Contains(b.Alias, ident)
	})
	if ix < 0 {
		return Builtin{}, fmt.Errorf("%s undefined builtin", ident)
	}
	return vs[ix], nil
}

func Lookup(ident string) (BuiltinFunc, error) {
	fn, err := Get(ident)
	if err != nil {
		return nil, err
	}
	return fn.Make(), nil
}

func List() []Builtin {
	vs := maps.Values(Registry)
	return slices.Collect(vs)
}

type Evaluable interface {
	Eval() value.Value
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

type deferrableValue struct {
	value.Value
}

func deferValue(val value.Value) value.Value {
	return deferrableValue{
		Value: val,
	}
}

func (d deferrableValue) Eval() value.Value {
	if e, ok := d.Value.(Evaluable); ok {
		return e.Eval()
	}
	return value.ErrValue
}

type Param struct {
	Name       string
	Desc       string
	Type       string
	Mode       value.ValueKind
	Optional   bool
	Variadic   bool
	Deferrable bool

	DefaultValue value.Value
}

func (p Param) Valid(val value.Value) bool {
	kd := val.Kind()
	return kd&p.Mode != 0
}

func (p Param) Convert(val value.Value) value.Value {
	if p.Deferrable {
		return deferValue(val)
	}
	e, ok := val.(Evaluable)
	if !ok {
		return value.ErrValue
	}
	val = e.Eval()
	if !p.Valid(val) {
		if value.IsError(val) {
			return val
		}
		return value.ErrValue
	}
	if value.IsArray(val) {
		arr, ok := val.(value.Array)
		if !ok {
			return val
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
				newArgs = append(newArgs, ret)
			}
			pix++
		}
		for _, p := range params[pix:] {
			if !p.Optional && !p.Variadic && len(newArgs) < len(params) {
				return value.ErrValue
			}
		}
		val := do(newArgs)
		if e, ok := val.(Evaluable); ok {
			return e.Eval()
		}
		return val
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

func OptDefault(p Param, val value.Value) Param {
	p.Optional = true
	return p
}

func Var(p Param) Param {
	p.Variadic = true
	return p
}

func Deferrable(p Param) Param {
	p.Deferrable = true
	return p
}

func asFloat(arg value.Value) float64 {
	v, _ := value.CastToFloat(arg)
	return float64(v)
}

func asFloatArray(args []value.Value) []float64 {
	arr := value.Collect[float64](args, func(v value.Value) (float64, bool) {
		if value.IsError(v) {
			return 0, false
		}
		return asFloat(v), true
	})
	return arr
}

func asString(arg value.Value) string {
	v, _ := value.CastToText(arg)
	return string(v)
}

func asBool(arg value.Value) bool {
	return value.True(arg)
}

func asTime(arg value.Value) time.Time {
	v, _ := value.CastToDate(arg)
	return time.Time(v)
}
