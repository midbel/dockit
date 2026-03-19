package grid

import (
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func NewContext(ctx value.Context) value.Context {
	if _, ok := ctx.(*evalContext); ok {
		return ctx
	}
	return createEvalContext(ctx)
}

func EnclosedContext(parent, child value.Context) value.Context {
	if e, ok := parent.(*evalContext); ok {
		x := &evalContext{
			stack: e.stack.Push(child),
		}
		return x
	}
	x := emptyEvalContext()
	x.push(parent)
	x.push(child)
	return x
}

type rowContext struct {
	rows []value.ScalarValue
}

func RowContext(rs []value.ScalarValue) value.Context {
	return rowContext{
		rows: rs,
	}
}

func (c rowContext) Resolve(string) value.Value {
	return value.ErrName
}

func (c rowContext) At(pos layout.Position) value.Value {
	if pos.Column < 0 || pos.Column >= int64(len(c.rows)) {
		return value.ErrNA
	}
	if pos.Line != 1 {
		return value.ErrNA
	}
	return c.rows[pos.Column-1]
}

func (c rowContext) Range(start, end layout.Position) value.Value {
	return value.ErrName
}

type sheetContext struct {
	view View
}

func SheetContext(view View) value.Context {
	return sheetContext{
		view: view,
	}
}

func (c sheetContext) ReadOnly() value.Context {
	return value.ReadOnly(c)
}

func (c sheetContext) Resolve(string) value.Value {
	return value.ErrName
}

func (c sheetContext) Range(start, end layout.Position) value.Value {
	if start.Sheet == "" || start.Sheet == c.view.Name() {
		rg := layout.NewRange(start, end)
		return ArrayView(NewBoundedView(c.view, rg))
	}
	return value.ErrRef
}

func (c sheetContext) At(pos layout.Position) value.Value {
	if pos.Sheet == "" || pos.Sheet == c.view.Name() {
		pos.Sheet = ""
		cell, err := c.view.Cell(pos)
		if err != nil || cell == nil {
			return value.ErrRef
		}
		if f := cell.Formula(); f != nil {
			return f
		}
		return cell.Value()
	}
	return value.ErrRef
}

func (c sheetContext) SetValue(pos layout.Position, val value.Value) error {
	mv, ok := c.view.(MutableView)
	if !ok {
		return ErrMutate
	}
	res, ok := val.(value.ScalarValue)
	if !ok {
		return ErrType
	}
	return mv.SetValue(pos, res)
}

func (c sheetContext) SetFormula(pos layout.Position, val value.Formula) error {
	mv, ok := c.view.(MutableView)
	if !ok {
		return ErrMutate
	}
	res, ok := val.(value.Formula)
	if !ok {
		return ErrType
	}
	return mv.SetFormula(pos, res)
}

func (c sheetContext) SetRange(start, end layout.Position, val value.Value) error {
	return nil
}

func (c sheetContext) SetRangeFormula(start, end layout.Position, val value.Value) error {
	return nil
}

type fileContext struct {
	file File
}

func FileContext(file File) value.Context {
	return fileContext{
		file: file,
	}
}

func (c fileContext) ReadOnly() value.Context {
	return value.ReadOnly(c)
}

func (c fileContext) Resolve(name string) value.Value {
	return value.ErrName
}

func (c fileContext) At(pos layout.Position) value.Value {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return value.ErrRef
	}
	return SheetContext(sh).At(pos)
}

func (c fileContext) Range(start, end layout.Position) value.Value {
	if start.Sheet != end.Sheet {
		return value.ErrNA
	}
	sh, err := c.sheet(start.Sheet)
	if err != nil {
		return value.ErrRef
	}
	return SheetContext(sh).Range(start, end)
}

func (c fileContext) SetValue(pos layout.Position, val value.Value) error {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return err
	}
	mv, ok := sh.(MutableView)
	if !ok {
		return ErrMutate
	}
	res, ok := val.(value.ScalarValue)
	if !ok {
		return ErrType
	}
	return mv.SetValue(pos, res)
}

func (c fileContext) SetFormula(pos layout.Position, val value.Formula) error {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return err
	}
	mv, ok := sh.(MutableView)
	if !ok {
		return ErrMutate
	}
	res, ok := val.(value.Formula)
	if !ok {
		return ErrType
	}
	return mv.SetFormula(pos, res)
}

func (c fileContext) SetRange(start, end layout.Position, val value.Value) error {
	return nil
}

func (c fileContext) SetRangeFormula(start, end layout.Position, val value.Value) error {
	return nil
}

func (c fileContext) sheet(name string) (View, error) {
	if name == "" {
		return c.file.ActiveSheet()
	}
	return c.file.Sheet(name)
}

type evalContext struct {
	stack *ds.Linked[value.Context]
}

func emptyEvalContext() *evalContext {
	return &evalContext{
		stack: ds.EmptyList[value.Context](),
	}
}

func createEvalContext(ctx value.Context) *evalContext {
	stack := ds.EmptyList[value.Context]().Push(ctx)
	return &evalContext{
		stack: stack,
	}
}

func (c *evalContext) ReadOnly() value.Context {
	return value.ReadOnly(c)
}

func (c *evalContext) Resolve(ident string) value.Value {
	var ret value.Value
	c.stack.Each(func(ctx value.Context) bool {
		ret = ctx.Resolve(ident)
		if ret == value.ErrName {
			return true
		}
		return value.IsError(ret)
	})
	return c.normalizeValue(ret)
}

func (c *evalContext) At(pos layout.Position) value.Value {
	var ret value.Value
	c.stack.Each(func(ctx value.Context) bool {
		ret = ctx.At(pos)
		if f, ok := ret.(value.Formula); ok {
			val, err := Eval(f, c)
			if err != nil {
				ret = value.ErrValue
			} else {
				ret = val
			}
		}
		if ret == value.ErrRef {
			return true
		}
		return value.IsError(ret)
	})
	return c.normalizeValue(ret)
}

func (c *evalContext) Range(start, end layout.Position) value.Value {
	var ret value.Value
	c.stack.Each(func(ctx value.Context) bool {
		ret = ctx.Range(start, end)
		if ret == value.ErrRef {
			return true
		}
		return value.IsError(ret)
	})
	return c.normalizeValue(ret)
}

func (c *evalContext) normalizeValue(val value.Value) value.Value {
	if val == nil {
		return value.ErrValue
	}
	return val
}

func (c *evalContext) push(ctx value.Context) {
	c.stack.Push(ctx)
}
