package eval

import (
	"errors"
	"fmt"
	"io"

	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

var (
	ErrSupported = errors.New("not supported")
	ErrEmpty     = errors.New("empty context")
	ErrMutate    = errors.New("context is not mutable")
)

type EngineConfig struct {
	Print struct {
		Cols  int
		Rows  int
		Debug bool
	}

	Formating struct {
		Date        string
		Bool        string
		Number      string
		ThousandSep rune
		DecimalSep  rune
	}

	ContextDir string

	Stdout io.Writer
	Stderr io.Writer
	Trace  bool

	Csv struct {
		Quoted bool
		Comma  rune
	}
}

func (c EngineConfig) Printer() Printer {
	if c.Print.Debug {
		return DebugValue(c.Stdout, c.Print.Rows, c.Print.Cols)
	}
	p := PrintValue(c.Stdout, c.Print.Rows, c.Print.Cols)
	if x, ok := p.(valuePrinter); ok {
		var vs valueFormatter
		if f, err := ParseNumberFormatter(c.Formating.Number); err == nil {
			vs.number = f
		}
		x.format = vs
		return x
	}
	return p
}

type EngineContext struct {
	ctx          *grid.ScopedContext
	config       EngineConfig
	currentValue value.Value
}

func NewEngineContext() *EngineContext {
	var cfg EngineConfig

	cfg.Formating.Number = defaultNumberFormatPattern
	cfg.Formating.ThousandSep = ','
	cfg.Formating.DecimalSep = '.'

	eg := EngineContext{
		ctx:    new(grid.ScopedContext),
		config: cfg,
	}
	return &eg
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
