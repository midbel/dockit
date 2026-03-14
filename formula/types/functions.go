package types

import (
	"github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/value"
)

type Function struct {
	name string
	fn   builtins.BuiltinFunc
}

func NewFunction(name string, fn builtins.BuiltinFunc) value.FunctionValue {
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

func (f Function) Call(args []value.Arg, ctx value.Context) value.Value {
	var values []value.Value
	for i := range args {
		a, _ := args[i].Eval(ctx)
		if value.IsError(a) {
			return a
		}
		values = append(values, a)
	}
	return f.fn(values)
}
