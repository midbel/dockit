package types

import (
	"fmt"

	"github.com/midbel/dockit/value"
)

type Reducer struct {
	name string
	fn   ReducerFunc
}

func NewReducer(name string, fn ReducerFunc) value.FunctionValue {
	return Reducer{
		name: name,
		fn:   fn,
	}
}

func (Reducer) Kind() value.ValueKind {
	return value.KindFunction
}

func (f Reducer) String() string {
	return f.name
}

func (f Reducer) Call(args []value.Arg, ctx value.Context) (value.Value, error) {
	if len(args) != 1 {
		return ErrValue, fmt.Errorf("%s only accepts one argument", f.name)
	}
	p, ok := args[0].(interface {
		asFilter(value.Context) (*value.Filter, bool, error)
	})
	if !ok {
		return ErrNA, fmt.Errorf("argument can not be used as argument")
	}
	if src, ok, err := p.asFilter(ctx); err == nil {
		if ok {
			return f.fn(src.Predicate, src.Value)
		}
		val, err := args[0].Eval(ctx)
		if err != nil {
			return nil, err
		}
		var p truePredicate
		return f.fn(p, val)
	} else {
		return ErrNA, err
	}
}

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