package builtins

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/midbel/dockit/value"
)

type BuiltinFunc func([]value.Value) (value.Value, error)

var (
	ErrArity = errors.New("invalid number of arguments")
	ErrType  = errors.New("invalid type")
)

func Fn(sig Signature, impl BuiltinFunc) BuiltinFunc {
	return impl
}

type Signature struct {
	params []Param
}

type Param struct {
	Type     string
	Mode     value.ValueKind
	Optional bool
	Variadic bool
}

func Scalar(k string) Param {
	return Param{
		Type: k,
		Mode: value.KindScalar,
	}
}

func Array(k string) Param {
	return Param{
		Type: k,
		Mode: value.KindArray,
	}
}

func ScalarArray(k string) Param {
	return Param{
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

func Each(args []value.Value, fn func(value.Value) error) error {
	for _, a := range args {
		if value.IsScalar(a) {
			if err := fn(a); err != nil {
				return err
			}
		} else if value.IsArray(a) {
			var (
				arr = a.(value.Array)
				dat []value.Value
			)
			for v := range arr.Values() {
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

var Registry = map[string]BuiltinFunc{
	"min":         Min,
	"max":         Max,
	"sum":         Sum,
	"avg":         Avg,
	"average":     Avg,
	"count":       Count,
	"round":       Round,
	"rounddown":   Floor,
	"roundup":     Ceil,
	"sqrt":        Sqrt,
	"abs":         Abs,
	"mod":         Mod,
	"pow":         Pow,
	"power":       Pow,
	"int":         Int,
	"floor":       Floor,
	"ceil":        Ceil,
	"sumif":       nil,
	"countif":     nil,
	"typeof":      TypeOf,
	"type":        TypeOf,
	"now":         Now,
	"today":       nil,
	"date":        nil,
	"year":        nil,
	"month":       nil,
	"day":         nil,
	"hour":        nil,
	"minute":      nil,
	"second":      nil,
	"weekday":     nil,
	"rand":        Rand,
	"isnumber":    IsNumber,
	"istext":      IsText,
	"isblank":     nil,
	"iserror":     nil,
	"isna":        nil,
	"concat":      Concat,
	"concatenate": Concat,
	"left":        Left,
	"right":       Right,
	"mid":         Mid,
	"len":         Len,
	"upper":       Upper,
	"lower":       Lower,
	"substr":      Substr,
	"replace":     Replace,
	"trim":        nil,
	"split":       nil,
	"join":        nil,
	"proper":      nil,
	"search":      nil,
	"find":        nil,
	"substitute":  nil,
	"text":        nil,
	"value":       nil,
	"textjoin":    nil,
	"if":          If,
	"iferror":     nil,
	"ifs":         nil,
	"ifna":        nil,
	"and":         And,
	"or":          Or,
	"xor":         Xor,
	"not":         Not,
	"lock":        Lock,
	"unlock":      Unlock,
	"index":       nil,
	"match":       nil,
	"vlookup":     nil,
	"hlookup":     nil,
	"xlookup":     nil,
	"offset":      nil,
	"choose":      nil,
}

func Lookup(ident string) (BuiltinFunc, error) {
	fn, ok := Registry[strings.ToLower(ident)]
	if !ok {
		return nil, fmt.Errorf("%s: undefined builtin", ident)
	}
	return fn, nil
}

func Now(args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, ErrArity
	}
	n := time.Now()
	return value.Date(n), nil
}

func Rand(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Lock(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Unlock(args []value.Value) (value.Value, error) {
	return nil, nil
}
