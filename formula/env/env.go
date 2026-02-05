package env

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrUndefined = errors.New("undefined identifier")
	ErrAvailable = errors.New("not available")
)

type Builtin interface {
	Call(args []value.Value) (value.Value, error)
	Arity() int
	Variadic() bool
}

type Environment struct {
	values map[string]value.Value

	currentVal value.Value
}

func Empty() *Environment {
	ctx := Environment{
		values: make(map[string]value.Value),
	}
	return &ctx
}

func (c *Environment) Resolve(ident string) (value.Value, error) {
	v, ok := c.values[ident]
	if ok {
		return v, nil
	}
	if x, ok := c.currentVal.(value.ObjectValue); ok {
		v, err := x.Get(ident)
		if err == nil {
			return v, nil
		}
	}
	return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
}

func (c *Environment) SetDefault(val value.Value) {
	c.currentVal = val
}

func (c *Environment) Default() value.Value {
	return c.currentVal
}

func (c *Environment) Define(ident string, val value.Value) {
	c.values[ident] = val
}

func (c *Environment) At(_ layout.Position) (value.Value, error) {
	return nil, ErrAvailable
}

func (c *Environment) Range(_, _ layout.Position) (value.Value, error) {
	return nil, ErrAvailable
}
