package eval

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/midbel/dockit/format"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var (
	ErrSupported = errors.New("not supported")
	ErrEmpty     = errors.New("empty context")
	ErrMutate    = errors.New("context is not mutable")
)

type EngineConfig struct {
	registry *ds.Trie[any]
}

func NewConfig() *EngineConfig {
	c := EngineConfig{
		registry: ds.NewTrie[any](),
	}
	c.SetDefaults()
	return &c
}

func (c *EngineConfig) Printer() (Printer, error) {
	debug, _ := c.registry.Get(slx.Make("print", "debug"))
	cols, _ := c.registry.Get(slx.Make("print", "cols"))
	rows, _ := c.registry.Get(slx.Make("print", "rows"))

	maxCols, ok := cols.(float64)
	if !ok {
		return nil, fmt.Errorf("columns should be a number")
	}
	maxRows, ok := rows.(float64)
	if !ok {
		return nil, fmt.Errorf("rows should be a number")
	}
	if d, ok := debug.(bool); ok && d {
		return DebugValue(os.Stdout, int(maxRows), int(maxCols)), nil
	}
	return PrintValue(os.Stdout, int(maxRows), int(maxCols)), nil
}

func (c *EngineConfig) Formatter() (format.Formatter, error) {
	vf := format.FormatValue()
	if num, ok := c.registry.Get(slx.Make("format", "number")); ok {
		str, ok := num.(string)
		if !ok {
			return nil, fmt.Errorf("number pattern should be a literal")
		}
		if err := vf.Number(str); err != nil {
			return nil, err
		}
	}
	if date, ok := c.registry.Get(slx.Make("format", "date")); ok {
		str, ok := date.(string)
		if !ok {
			return nil, fmt.Errorf("date pattern should be a literal")
		}
		if err := vf.Date(str); err != nil {
			return nil, err
		}
	}
	if mode, ok := c.registry.Get(slx.Make("format", "bool")); ok {
		str, ok := mode.(string)
		if !ok {
			return nil, fmt.Errorf("boolean pattern should be a literal")
		}
		if err := vf.Bool(str); err != nil {
			return nil, err
		}
	}
	return vf, nil
}

func (c *EngineConfig) Set(ident []string, val any) error {
	if val == nil {
		return nil
	}
	c.registry.Register(ident, val)
	return nil
}

func (c *EngineConfig) SetDefaults() {
	c.Set(slx.Make("context", "dir"), ".")
	c.Set(slx.Make("print", "debug"), false)
	c.Set(slx.Make("print", "cols"), maxCols)
	c.Set(slx.Make("print", "rows"), maxRows)
	c.Set(slx.Make("format", "number"), format.DefaultNumberPattern)
	c.Set(slx.Make("format", "date"), format.DefaultDatePattern)
	c.Set(slx.Make("format", "bool"), "")
}

type EngineContext struct {
	ctx          *grid.ScopedContext
	currentValue value.Value

	printer    Printer
	formatter  format.Formatter
	contextDir string
}

func NewEngineContext() *EngineContext {
	eg := EngineContext{
		ctx: new(grid.ScopedContext),
	}
	return &eg
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

	return nil
}

func (c *EngineContext) Print(v value.Value) {
	c.printer.Format(v, c.formatter)
}

func (c *EngineContext) Resolve(ident string) (value.Value, error) {
	if obj, ok := c.currentValue.(value.ObjectValue); ok {
		v, err := obj.Get(ident)
		if err == nil {
			return v, err
		}
	}
	return c.ctx.Resolve(ident)
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

func (c *EngineContext) Context() value.Context {
	return c.ctx
}

func (c *EngineContext) PushContext(ctx value.Context) {
	c.ctx.Push(ctx)
}

func (c *EngineContext) PushValue(val value.Value, ident string) (io.Closer, error) {
	sh, err := c.getActiveView(val, ident)
	if err != nil {
		return nil, err
	}
	ctx := grid.SheetContext(sh.View())
	if file, ok := val.(*types.File); ok {
		fc := grid.FileContext(file.File())
		ctx = grid.EvalContext(fc, ctx)
	}
	n := c.ctx.Len()
	c.ctx.Push(ctx)

	cf := func() {
		c.ctx.Truncate(n)
	}
	return closable(cf), nil
}

func (c *EngineContext) PushMutable(name string) (io.Closer, error) {
	sub, err := c.mutableView(name)
	if err != nil {
		return nil, err
	}
	if f, ok := c.currentValue.(*types.File); ok {
		fc := grid.FileContext(f.File())
		sub = grid.EvalContext(fc, sub)
	}
	n := c.ctx.Len()
	c.ctx.Push(sub)

	cf := func() {
		c.ctx.Truncate(n)
	}
	return closable(cf), nil
}

func (c *EngineContext) PushReadable(name string) (io.Closer, error) {
	sub, err := c.readableView(name)
	if err != nil {
		return nil, err
	}
	if f, ok := c.currentValue.(*types.File); ok {
		fc := grid.FileContext(f.File())
		sub = grid.EvalContext(fc, sub)
	}
	n := c.ctx.Len()
	c.ctx.Push(sub)

	cf := func() {
		c.ctx.Truncate(n)
	}
	return closable(cf), nil
}

func (c *EngineContext) readableView(name string) (value.Context, error) {
	sh, err := c.getActiveView(c.Default(), name)
	if err != nil {
		return nil, err
	}
	return value.ReadOnly(grid.SheetContext(sh.View())), nil
}

func (c *EngineContext) mutableView(name string) (value.Context, error) {
	sh, err := c.getActiveView(c.Default(), name)
	if err != nil {
		return nil, err
	}
	view, err := sh.Mutable()
	if err != nil {
		return nil, ErrReadOnly
	}
	return grid.SheetContext(view), nil
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

type closable func()

func (c closable) Close() error {
	c()
	return nil
}
