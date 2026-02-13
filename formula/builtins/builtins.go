package builtins

import (
	"errors"

	"github.com/midbel/dockit/value"
)

var ErrArity = errors.New("invalid number of arguments")

var Registry = map[string]func([]value.Value) (value.Value, error){
	"min":       Min,
	"max":       Max,
	"sum":       Sum,
	"avg":       Avg,
	"average":   Avg,
	"count":     Count,
	"round":     Round,
	"rounddown": Floor,
	"roundup":   Ceil,
	"sqrt":      Sqrt,
	"typeof":    TypeOf,
	"now":       Now,
	"rand":      Rand,
	"isnumber":  IsNumber,
	"istext":    IsText,
	"concat":    Concat,
	"left":      Left,
	"right":     Right,
	"mid":       Mid,
	"len":       Len,
	"upper":     Upper,
	"lower":     Lower,
	"substr":    Substr,
	"replace":   Replace,
	"if":        If,
	"and":       And,
	"or":        Or,
	"xor":       Xor,
	"not":       Not,
}

func Now(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Rand(args []value.Value) (value.Value, error) {
	return nil, nil
}
