package eval

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrSupported = errors.New("not supported")
	ErrEmpty     = errors.New("empty context")
	ErrMutate    = errors.New("context is not mutable")
)

type EngineContext struct {
	loaders      map[string]Loader
	env          *env.Environment
	currentValue value.Value

	printer    Printer
	formatter  format.Formatter
	contextDir string
}

func NewEngineContext() *EngineContext {
	return new(EngineContext)
}

func (c *EngineContext) Configure(cfg *EngineConfig) error {
	f, err := cfg.Formatter()
	if err != nil {
		return err
	}
	c.formatter = f

	p, err := cfg.Printer()
	if err != nil {
		return err
	}
	c.printer = p

	c.contextDir = cfg.ContextDir()

	return nil
}

func (c *EngineContext) Open(file string, opts LoaderOptions) (grid.File, error) {
	ext := filepath.Ext(file)
	loader, ok := c.loaders[ext]
	if !ok {
		return nil, fmt.Errorf("file %s can not be loaded!", ext)
	}
	file = filepath.Join(c.contextDir, file)
	return loader.Open(file, opts)
}

func (c *EngineContext) Print(v value.Value) {
	c.printer.Format(v, c.formatter)
}

func (c *EngineContext) Resolve(ident string) value.Value {
	if obj, ok := c.currentValue.(value.ObjectValue); ok {
		val := obj.Get(ident)
		if !value.IsError(val) {
			return val
		}
	}
	return c.env.Resolve(ident)
}

func (c *EngineContext) Define(ident string, value value.Value) {
	c.env.Define(ident, value)
}

func (c *EngineContext) Default() value.Value {
	return c.currentValue
}

func (c *EngineContext) CurrentActiveView() *types.View {
	switch c := c.currentValue.(type) {
	case *types.View:
		return c
	case *types.File:
		active, err := c.Active()
		if err != nil {
			return nil
		}
		v, ok := active.(*types.View)
		if ok {
			return v
		}
		return nil
	default:
		return nil
	}
}

func (c *EngineContext) SetDefault(val value.Value) {
	c.currentValue = val
}

func (c *EngineContext) At(pos layout.Position) value.Value {
	sh, err := c.getActiveView(c.Default(), pos.Sheet)
	if err != nil {
		return value.ErrNA
	}
	return sh.At(pos)
}

func (c *EngineContext) Range(start, end layout.Position) value.Value {
	sh, err := c.getActiveView(c.Default(), start.Sheet)
	if err != nil {
		return value.ErrNA
	}
	return sh.Range(start, end)
}

func (c *EngineContext) setEnv(environ *env.Environment) {
	c.env = environ
}

func (c *EngineContext) getActiveView(val value.Value, name string) (*types.View, error) {
	switch v := val.(type) {
	case *types.View:
		return v, nil
	case *types.File:
		return c.getViewFromFile(v, name)
	default:
		return nil, fmt.Errorf("%s: view can not be found", name)
	}
}

func (c *EngineContext) getViewFromFile(file *types.File, name string) (*types.View, error) {
	var (
		sheet value.Value
		err   error
	)
	if name == "" {
		sheet, err = file.Active()
	} else {
		sheet, err = file.Sheet(name)
	}
	if err != nil {
		return nil, err
	}
	tv, ok := sheet.(*types.View)
	if !ok {
		return nil, ErrValue
	}
	return tv, nil
}
