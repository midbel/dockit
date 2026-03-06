package builtins

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/midbel/dockit/value"
)

type BuiltinFunc func([]value.Value) (value.Value, error)

var ErrArity = errors.New("invalid number of arguments")

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
	"abs":         nil,
	"mod":         nil,
	"pow":         nil,
	"int":         nil,
	"floor":       nil,
	"ceil":        nil,
	"sumif":       nil,
	"countif":     nil,
	"typeof":      TypeOf,
	"now":         Now,
	"today":       nil,
	"date":        nil,
	"year":        nil,
	"month":       nil,
	"day":         nil,
	"rand":        Rand,
	"isnumber":    IsNumber,
	"istext":      IsText,
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
	"if":          If,
	"and":         And,
	"or":          Or,
	"xor":         Xor,
	"not":         Not,
	"lock":        Lock,
	"unlock":      Unlock,
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
