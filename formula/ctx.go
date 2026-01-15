package formula

import (
	"errors"
	"fmt"
	"os"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrUndefined = errors.New("undefined identifier")
	ErrAvailable = errors.New("not available")
)

type BuiltinFunc func([]value.Value) (value.Value, error)

type envValue struct{}

func (*envValue) Kind() value.ValueKind {
	return value.KindObject
}

func (*envValue) String() string {
	return "env"
}

func (v *envValue) Get(name string) (value.ScalarValue, error) {
	str := os.Getenv(name)
	return Text(str), nil
}

type Environment struct {
	values map[string]value.Value
	parent value.Context
}

func Enclosed(parent value.Context) *Environment {
	ctx := Environment{
		values: make(map[string]value.Value),
		parent: parent,
	}
	return &ctx
}

func Empty() *Environment {
	return Enclosed(nil)
}

func (c *Environment) Resolve(ident string) (value.Value, error) {
	v, ok := c.values[ident]
	if ok {
		return v, nil
	}
	if c.parent == nil {
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
	return c.parent.Resolve(ident)
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
