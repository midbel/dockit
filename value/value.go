package value

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
)

type ValueKind int8

const (
	KindScalar ValueKind = 1 << iota
	KindError
	KindArray
	KindObject
	KindFunction
)

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
