package types

import (
	"github.com/midbel/dockit/value"
)

type BuiltinFunc func([]value.Value) (value.Value, error)

type Function struct {
	name string
	fn   BuiltinFunc
}

func NewFunction(name string, fn BuiltinFunc) value.FunctionValue {
	return Function{
		name: name,
		fn:   fn,
	}
}

func (Function) Type() string {
	return "function"
}

func (Function) Kind() value.ValueKind {
	return value.KindFunction
}

func (f Function) String() string {
	return f.name
}

func (f Function) Call(args []value.Arg, ctx value.Context) (value.Value, error) {
	var values []value.Value
	for i := range args {
		a, err := args[i].Eval(ctx)
		if err != nil {
			return nil, err
		}
		values = append(values, a)
	}
	return f.fn(values)
}
