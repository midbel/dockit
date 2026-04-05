package eval

import (
	"fmt"
	"path/filepath"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type EngineContext struct {
	loaders      map[string]Loader
	env          *env.Environment
	currentValue value.Value

	printer    Printer
	formatter  format.Formatter
	contextDir string
	config     *EngineConfig
}

func NewEngineContext() *EngineContext {
	return new(EngineContext)
}

func (c *EngineContext) Sub(val value.Value) *EngineContext {
	if val == c.currentValue {
		return c
	}
	x := *c
	x.currentValue = val
	return &x
}

func (c *EngineContext) Configure(cfg *EngineConfig) error {
	c.config = cfg
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

func (c *EngineContext) GetOption(key []string) any {
	return c.config.Get(key)
}

func (c *EngineContext) GetOptionString(key []string) string {
	v := c.config.Get(key)
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
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

func (c *EngineContext) Print(v value.Value) error {
	if s, ok := v.(interface{ Sync() error }); ok {
		if err := s.Sync(); err != nil {
			return err
		}
	}
	c.printer.Format(v, c.formatter)
	return nil
}

func (c *EngineContext) Export(v value.Value) error {
	if s, ok := v.(interface{ Sync() error }); ok {
		if err := s.Sync(); err != nil {
			return err
		}
	}
	return nil
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

func (c *EngineContext) SetDefault(val value.Value) {
	c.currentValue = val
}

func (c *EngineContext) CurrentActiveView() *types.View {
	v, _ := c.getView("")
	return v
}

func (c *EngineContext) At(pos layout.Position) value.Value {
	sh, err := c.getView(pos.Sheet)
	if err != nil {
		return value.ErrNA
	}
	return sh.At(pos)
}

func (c *EngineContext) SetAt(pos layout.Position, val value.Value) error {
	sh, err := c.getView(pos.Sheet)
	if err != nil {
		return err
	}
	return sh.SetAt(pos, val)
}

func (c *EngineContext) Range(start, end layout.Position) value.Value {
	if start.Sheet != end.Sheet {
		return value.ErrName
	}
	sh, err := c.getView(start.Sheet)
	if err != nil {
		return value.ErrNA
	}
	return sh.Range(start, end)
}

func (c *EngineContext) SetRange(start, end layout.Position, val value.Value) error {
	sh, err := c.getView(start.Sheet)
	if err != nil {
		return err
	}
	return sh.SetRange(start, end, val)
}

func (c *EngineContext) setEnv(environ *env.Environment) {
	c.env = environ
}

func (c *EngineContext) getView(name string) (*types.View, error) {
	if f, ok := c.Default().(*types.File); ok {
		return c.getViewFromFile(f, name)
	}
	return nil, fmt.Errorf("%s: view can not be found", name)
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
		err = types.ErrType
	}
	return tv, err
}
