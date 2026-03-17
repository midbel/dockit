package grid

import (
	"slices"

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
	var (
		base   value.Context
		scoped value.Context
	)
	if e, ok := parent.(*evalContext); ok {
		base = e.inner
	} else {
		base = parent
	}
	if ps, ok := base.(*ScopedContext); ok {
		ps := ps.Clone()
		ps.Push(child)
		scoped = ps
	} else {
		scoped = NewScopedContext(base, child)
	}
	return NewContext(scoped)
}

type ScopedContext struct {
	parent *ScopedContext
	ctx value.Context
}

func NewScopedContext(others ...value.Context) value.Context {
	es := ScopedContext(slices.Clone(others))
	return &es
}

func (ec *ScopedContext) Clone() *ScopedContext {
	var (
		vs  = slices.Clone(*ec)
		ctx = ScopedContext(vs)
	)
	return &ctx
}

func (ec *ScopedContext) Push(ctx value.Context) {
	*ec = append(*ec, ctx)
}

func (ec *ScopedContext) Pop() {
	if n := len(*ec); n >= 1 {
		*ec = (*ec)[:n-1]
	}
}

func (ec *ScopedContext) Len() int {
	return len(*ec)
}

func (ec *ScopedContext) Truncate(n int) {
	z := len(*ec)
	if n >= z {
		n = z
	}
	*ec = (*ec)[:n]
}

func (ec *ScopedContext) ReadOnly() value.Context {
	return value.ReadOnly(ec)
}

func (ec *ScopedContext) Resolve(name string) value.Value {
	for i := len(*ec) - 1; i >= 0; i-- {
		v := (*ec)[i].Resolve(name)
		if !value.IsError(v) {
			return v
		}
	}
	return value.ErrName
}

func (ec *ScopedContext) At(pos layout.Position) value.Value {
	for i := len(*ec) - 1; i >= 0; i-- {
		v := (*ec)[i].At(pos)
		if !value.IsError(v) {
			return v
		}
	}
	return value.ErrRef
}

func (ec *ScopedContext) Range(start, end layout.Position) value.Value {
	for i := len(*ec) - 1; i >= 0; i-- {
		v := (*ec)[i].Range(start, end)
		if !value.IsError(v) {
			return v
		}
	}
	return value.ErrRef
}

func (ec *ScopedContext) Define(ident string, val value.Value) {
	ctx := ec.top()
	if ctx == nil {
		return
	}
	if e, ok := ctx.(interface{ Define(string, value.Value) }); ok {
		e.Define(ident, val)
	}
}

func (ec *ScopedContext) SetValue(pos layout.Position, val value.Value) error {
	ctx := ec.top()
	if ctx == nil {
		return ErrEmpty
	}
	if mc, ok := ctx.(value.MutableContext); ok {
		return mc.SetValue(pos, val)
	}
	return nil
}

func (ec *ScopedContext) SetFormula(pos layout.Position, val value.Formula) error {
	ctx := ec.top()
	if ctx == nil {
		return ErrEmpty
	}
	if mc, ok := ctx.(value.MutableContext); ok {
		return mc.SetFormula(pos, val)
	}
	return nil
}

func (ec *ScopedContext) SetRange(start, end layout.Position, val value.Value) error {
	return nil
}

func (ec *ScopedContext) SetRangeFormula(start, end layout.Position, val value.Value) error {
	return nil
}

func (ec *ScopedContext) top() value.Context {
	n := len(*ec)
	if n > 0 {
		return (*ec)[n-1]
	}
	return nil
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
	rg := layout.NewRange(start, end)
	return ArrayView(NewBoundedView(c.view, rg))
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
	inner value.Context
}

func createEvalContext(ctx value.Context) value.Context {
	return &evalContext{
		inner: ctx,
	}
}

func (c *evalContext) Resolve(ident string) value.Value {
	return c.inner.Resolve(ident)
}

func (c *evalContext) At(pos layout.Position) value.Value {
	v := c.inner.At(pos)
	if f, ok := v.(value.Formula); ok {
		val, err := Eval(f, c)
		if err != nil {
			v = value.ErrValue
		} else {
			v = val
		}
	}
	return v
}

func (c *evalContext) Range(start, end layout.Position) value.Value {
	return c.inner.Range(start, end)
}
