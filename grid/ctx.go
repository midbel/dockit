package grid

import (
	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type viewContext struct {
	view View
}

func (c *viewContext) Resolve(ident string) (value.Value, error) {
	return nil, nil
}

func (c *viewContext) At(_ layout.Position) (value.Value, error) {
	return nil, ErrSupported
}

func (c *viewContext) Range(_, _ layout.Position) (value.Value, error) {
	return nil, ErrSupported
}

func (c *viewContext) ResolveFunc(_ string) (formula.BuiltinFunc, error) {
	return nil, ErrSupported
}

type fileContext struct {
	file File
}

func (c *fileContext) Resolve(ident string) (value.Value, error) {
	return nil, nil
}

func (c *fileContext) At(_ layout.Position) (value.Value, error) {
	return nil, ErrSupported
}

func (c *fileContext) Range(_, _ layout.Position) (value.Value, error) {
	return nil, ErrSupported
}

func (c *fileContext) ResolveFunc(_ string) (formula.BuiltinFunc, error) {
	return nil, ErrSupported
}
