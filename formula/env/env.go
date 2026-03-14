package env

import (
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type Environment struct {
	values map[string]value.Value
}

func Empty() *Environment {
	ctx := Environment{
		values: make(map[string]value.Value),
	}
	return &ctx
}

func (c *Environment) Resolve(ident string) value.Value {
	v, ok := c.values[ident]
	if ok {
		return v
	}
	return value.ErrName
}

func (c *Environment) Define(ident string, val value.Value) {
	c.values[ident] = val
}

func (c *Environment) At(_ layout.Position) value.Value {
	return value.ErrRef
}

func (c *Environment) Range(_, _ layout.Position) value.Value {
	return value.ErrRef
}
