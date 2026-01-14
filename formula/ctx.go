package formula

import (
	"errors"
	"fmt"
	"os"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var ErrUndefined = errors.New("undefined identifier")

type BuiltinFunc func([]value.Value) (value.Value, error)

type EnvContext struct{}

func (c *EnvContext) Resolve(ident string) (value.Value, error) {
	str := os.Getenv(ident)
	return Text(str), nil
}

func (c *EnvContext) At(_ layout.Position) (value.Value, error) {
	return nil, nil
}

func (c *EnvContext) Range(_, _ layout.Position) (value.Value, error) {
	return nil, nil
}

type VarsContext struct {
	values map[string]value.Value
	parent value.Context
}

func Enclosed(parent value.Context) value.Context {
	ctx := VarsContext{
		values: make(map[string]value.Value),
		parent: parent,
	}
	return &ctx
}

func Empty() value.Context {
	return Enclosed(nil)
}

func (c *VarsContext) Resolve(ident string) (value.Value, error) {
	v, ok := c.values[ident]
	if ok {
		return v, nil
	}
	if c.parent == nil {
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
	return c.parent.Resolve(ident)
}

func (c *VarsContext) At(_ layout.Position) (value.Value, error) {
	return nil, nil
}

func (c *VarsContext) Range(_, _ layout.Position) (value.Value, error) {
	return nil, nil
}
