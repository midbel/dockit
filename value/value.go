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
