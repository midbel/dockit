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
)

type Value interface {
	Kind() ValueKind
	fmt.Stringer
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
	Get(string) (ScalarValue, error)
}

type CastableValue interface {
	ToString() ScalarValue
	ToBool() ScalarValue
	ToFloat() ScalarValue
	// ToDate() ScalarValue
}
