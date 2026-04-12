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
	KindInspectable
)

const (
	TypeNumber = "number"
	TypeText   = "text"
	TypeBool   = "boolean"
	TypeDate   = "date"
	TypeBlank  = "blank"
	TypeArray  = "array"
	TypeError  = "error"
	TypeAny    = "any"
)

func IsComparable(v Value) bool {
	_, ok := v.(Comparable)
	return ok
}

func IsNumber(v Value) bool {
	_, ok := v.(Float)
	return ok
}
func IsText(v Value) bool {
	_, ok := v.(Text)
	return ok
}

func IsScalar(v Value) bool {
	if v == nil {
		return false
	}
	return v.Kind() == KindScalar
}

func IsObject(v Value) bool {
	if v == nil {
		return false
	}
	return v.Kind() == KindObject
}

func IsArray(v Value) bool {
	if v == nil {
		return false
	}
	return v.Kind() == KindArray
}

func IsError(v Value) bool {
	if v == nil {
		return false
	}
	return v.Kind() == KindError && v.Type() == TypeError
}

func IsBlank(v Value) bool {
	_, ok := v.(Blank)
	return ok
}

func Rows(rs ...[]ScalarValue) [][]ScalarValue {
	return rs
}

type Context interface {
	At(layout.Position) Value
	Range(layout.Position, layout.Position) Value
	Resolve(string) Value
}

type Formula interface {
	Value
	Eval(Context) Value
}

type Value interface {
	Kind() ValueKind
	Type() string
	fmt.Stringer
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
	Get(string) Value
}

type Predicate interface {
	Test(Context) bool
}

type CastableValue interface {
	ToString() ScalarValue
	ToBool() ScalarValue
	ToFloat() ScalarValue
	// ToDate() ScalarValue
}

func Add(left, right Value) Value {
	a, ok := left.(interface {
		Add(Value) ScalarValue
	})
	if !ok {
		return ErrValue
	}
	return a.Add(right)
}

func Sub(left, right Value) Value {
	a, ok := left.(interface {
		Sub(Value) ScalarValue
	})
	if !ok {
		return ErrValue
	}
	return a.Sub(right)
}

func Mul(left, right Value) Value {
	a, ok := left.(interface {
		Mul(Value) ScalarValue
	})
	if !ok {
		return ErrValue
	}
	return a.Mul(right)
}

func Div(left, right Value) Value {
	a, ok := left.(interface {
		Div(Value) ScalarValue
	})
	if !ok {
		return ErrValue
	}
	return a.Div(right)
}

func Pow(left, right Value) Value {
	a, ok := left.(interface {
		Pow(Value) ScalarValue
	})
	if !ok {
		return ErrValue
	}
	return a.Pow(right)
}

func Concat(left, right Value) Value {
	ls, err := CastToText(left)
	if err != nil {
		return ErrValue
	}
	rs, err := CastToText(right)
	if err != nil {
		return ErrValue
	}
	return Text(ls + rs)
}

func Eq(left, right Value) Value {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue
	}
	ok, err := cmp.Equal(right)
	if err != nil {
		return ErrValue
	}
	return Boolean(ok)
}

func Ne(left, right Value) Value {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue
	}
	ok, err := cmp.Equal(right)
	if err != nil {
		return ErrValue
	}
	return Boolean(!ok)
}

func Lt(left, right Value) Value {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue
	}
	ok, err := cmp.Less(right)
	if err != nil {
		return ErrValue
	}
	return Boolean(ok)
}

func Le(left, right Value) Value {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(ok)
	}
	ok, err = cmp.Less(right)
	if err != nil {
		return ErrValue
	}
	return Boolean(ok)
}

func Gt(left, right Value) Value {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(!ok)
	}
	ok, err = cmp.Less(right)
	if err != nil {
		return ErrValue
	}
	return Boolean(!ok)
}

func Ge(left, right Value) Value {
	cmp, ok := left.(Comparable)
	if !ok {
		return ErrValue
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(ok)
	}
	ok, err = cmp.Less(right)
	if err != nil {
		return ErrValue
	}
	return Boolean(!ok)
}
