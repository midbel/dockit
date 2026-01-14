package formula

import (
	"errors"
	"fmt"
	"os"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var ErrUndefined = errors.New("undefined identifier")

type Context interface {
	At(layout.Position) (value.Value, error)
	Range(layout.Position, layout.Position) (value.Value, error)
	Resolve(ident string) (value.Value, error)
	ResolveFunc(ident string) (BuiltinFunc, error)
}

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

func (c *EnvContext) ResolveFunc(_ string) (BuiltinFunc, error) {
	return nil, nil
}

type VarsContext struct {
	values map[string]value.Value
	parent Context
}

func Enclosed(parent Context) Context {
	ctx := VarsContext{
		values: make(map[string]value.Value),
		parent: parent,
	}
	return &ctx
}

func Empty() Context {
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

func (c *VarsContext) ResolveFunc(_ string) (BuiltinFunc, error) {
	return nil, nil
}
