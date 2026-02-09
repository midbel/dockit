package value

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
)

var (
	ErrCompatible = errors.New("incompatible type")
	ErrCast       = errors.New("value can not be cast to target type")
)

type ValueKind int8

const (
	KindScalar ValueKind = 1 << iota
	KindError
	KindArray
	KindObject
	KindFunction
)

func IsComparable(v Value) bool {
	_, ok := v.(Comparable)
	return ok
}

func IsNumber(v Value) bool {
	_, ok := v.(Float)
	return ok
}

func IsScalar(v Value) bool {
	return v.Kind() == KindScalar
}

func IsObject(v Value) bool {
	return v.Kind() == KindObject
}

func IsArray(v Value) bool {
	return v.Kind() == KindArray
}

type Context interface {
	At(layout.Position) (Value, error)
	Range(layout.Position, layout.Position) (Value, error)
	Resolve(string) (Value, error)
}

type readonlyContext struct {
	inner Context
}

func ReadOnly(ctx Context) Context {
	return readonlyContext{
		inner: ctx,
	}
}

func (c readonlyContext) At(pos layout.Position) (Value, error) {
	return c.inner.At(pos)
}

func (c readonlyContext) Range(start, end layout.Position) (Value, error) {
	return c.inner.Range(start, end)
}

func (c readonlyContext) Resolve(ident string) (Value, error) {
	return c.inner.Resolve(ident)
}

type MutableContext interface {
	SetValue(layout.Position, Value) error
	SetFormula(layout.Position, Formula) error
	SetRange(layout.Position, layout.Position, Value) error
	SetRangeFormula(layout.Position, layout.Position, Value) error
}

type Formula interface {
	Eval(Context) (Value, error)
}

type Predicate interface {
	Test(ScalarValue) (bool, error)
}

type Value interface {
	Kind() ValueKind
	Type() string
	fmt.Stringer
}

type Filter struct {
	Predicate
	Value
}

type Comparable interface {
	Equal(Value) (bool, error)
	Less(Value) (bool, error)
}

type ScalarValue interface {
	Value
	Scalar() any
}

type ArrayValue interface {
	Value
	Dimension() layout.Dimension
	At(int, int) ScalarValue
}

var ErrProp = errors.New("undefined property")

type ObjectValue interface {
	Value
	Get(string) (Value, error)
}

type Arg interface {
	Eval(Context) (Value, error)
}

type FunctionValue interface {
	Value
	Call([]Arg, Context) (Value, error)
}

type CastableValue interface {
	ToString() ScalarValue
	ToBool() ScalarValue
	ToFloat() ScalarValue
	// ToDate() ScalarValue
}

var ErrOperation = errors.New("operation not supported")

func Add(left, right Value) (Value, error) {
	a, ok := left.(interface {
		Add(Value) (ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s + %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Add(right)
}

func Sub(left, right Value) (Value, error) {
	a, ok := left.(interface {
		Sub(Value) (ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s - %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Sub(right)
}

func Mul(left, right Value) (Value, error) {
	a, ok := left.(interface {
		Mul(Value) (ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s * %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Mul(right)
}

func Div(left, right Value) (Value, error) {
	a, ok := left.(interface {
		Div(Value) (ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s / %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Div(right)
}

func Pow(left, right Value) (Value, error) {
	a, ok := left.(interface {
		Pow(Value) (ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s ^ %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Pow(right)
}

func Concat(left, right Value) (Value, error) {
	ls, err := CastToText(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToText(right)
	if err != nil {
		return nil, err
	}
	return Text(ls + rs), nil
}

func Eq(left, right Value) (Value, error) {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	return Boolean(ok), err
}

func Ne(left, right Value) (Value, error) {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	return Boolean(!ok), err
}

func Lt(left, right Value) (Value, error) {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Less(right)
	return Boolean(!ok), err
}

func Le(left, right Value) (Value, error) {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(ok), nil
	}
	ok, err = cmp.Less(right)
	return Boolean(ok), err
}

func Gt(left, right Value) (Value, error) {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(!ok), nil
	}
	ok, err = cmp.Less(right)
	return Boolean(!ok), err
}

func Ge(left, right Value) (Value, error) {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(ok), nil
	}
	ok, err = cmp.Less(right)
	return Boolean(!ok), err
}
