package builtins

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/value"
)

var registry = map[string]Builtin{}

func Get(ident string) (Builtin, error) {
	fn, ok := registry[strings.ToLower(ident)]
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
	vs := maps.Values(registry)
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

func Object(name, desc string, k string) Param {
	return Param{
		Name: name,
		Desc: desc,
		Type: k,
		Mode: value.KindObject,
	}
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

func init() {
	registerBuiltins(condBuiltins)
	registerBuiltins(timeBuiltins)
	registerBuiltins(typeBuiltins)
	registerBuiltins(indexBuiltins)
	registerBuiltins(textBuiltins)
	registerBuiltins(numberBuiltins)
}

func registerBuiltins(list []Builtin) {
	for _, b := range list {
		registry[b.Name] = b
	}
}
