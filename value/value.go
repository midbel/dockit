package value

import (
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

type Predicate interface {
	Test(Context, ScalarValue) (bool, error)
}

type Value interface {
	Kind() ValueKind
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
