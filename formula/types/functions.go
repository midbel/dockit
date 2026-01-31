package types

import (
	"fmt"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/value"
)

type ReducerFunc func(value.Predicate, value.Value) (value.Value, error)

type BuiltinFunc func([]value.Value) (value.Value, error)

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

func (Reducer) Type() string {
	return "function"
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

type truePredicate struct{}

func (truePredicate) Test(value.ScalarValue) (bool, error) {
	return true, nil
}

type cmpPredicate struct {
	op.Op
	scalar value.ScalarValue
}

func (p cmpPredicate) Test(other value.ScalarValue) (bool, error) {
	c, ok := p.scalar.(value.Comparable)
	if !ok {
		return false, fmt.Errorf("value is not comparable")
	}
	var err error
	switch p.Op {
	case op.Eq:
		ok, err = c.Equal(other)
	case op.Ne:
		ok, err = c.Equal(other)
		ok = !ok
	case op.Lt:
		return c.Less(other)
	case op.Le:
		ok, err = c.Equal(other)
		if ok && err == nil {
			break
		}
		ok, err = c.Less(other)
	case op.Gt:
	case op.Ge:
		ok, err = c.Equal(other)
		if ok && err == nil {
			break
		}
	default:
	}
	return ok, err
}

func NewPredicate(oper op.Op, val value.Value) (value.Predicate, error) {
	scalar, ok := val.(value.ScalarValue)
	if !ok {
		return nil, fmt.Errorf("predicate can only operate on scalar value")
	}
	var p value.Predicate
	switch oper {
	case op.Eq, op.Ne, op.Lt, op.Le, op.Gt, op.Ge:
		p = cmpPredicate{
			Op:     oper,
			scalar: scalar,
		}
	default:
		return nil, fmt.Errorf("unsupported predicate type")
	}
	return p, nil
}
