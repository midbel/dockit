package value

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
)

var ErrUndefined = errors.New("undefined identifier")

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

type CastableValue interface {
	ToString() ScalarValue
	ToBool() ScalarValue
	ToFloat() ScalarValue
	// ToDate() ScalarValue
}

type Context struct {
	values map[string]Value
	parent *Context
}

func Enclosed(parent *Context) *Context {
	ctx := Context{
		values: make(map[string]Value),
		parent: parent,
	}
	return &ctx
}

func Empty() *Context {
	return Enclosed(nil)
}

func (c *Context) Resolve(ident string) (Value, error) {
	v, ok := c.values[ident]
	if ok {
		return v, nil
	}
	if c.parent == nil {
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
	return c.parent.Resolve(ident)
}
