package grid

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type ScopedContext []value.Context

func EvalContext(others ...value.Context) value.Context {
	es := ScopedContext(others)
	return &es
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

func (ec *ScopedContext) Resolve(name string) (value.Value, error) {
	for i := len(*ec) - 1; i >= 0; i-- {
		v, err := (*ec)[i].Resolve(name)
		if err == nil {
			return v, err
		}
	}
	return value.ErrValue, nil
}

func (ec *ScopedContext) At(pos layout.Position) (value.Value, error) {
	for i := len(*ec) - 1; i >= 0; i-- {
		v, err := (*ec)[i].At(pos)
		if err == nil {
			return v, err
		}
	}
	return value.ErrValue, nil
}

func (ec *ScopedContext) Range(start, end layout.Position) (value.Value, error) {
	for i := len(*ec) - 1; i >= 0; i-- {
		v, err := (*ec)[i].Range(start, end)
		if err == nil {
			return v, err
		}
	}
	return value.ErrValue, nil
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

func (c sheetContext) Resolve(name string) (value.Value, error) {
	return nil, ErrSupported
}

func (c sheetContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return nil, fmt.Errorf("cross sheet range not allowed")
	}
	var sh View
	if start.Sheet == "" || start.Sheet == c.view.Name() {
		sh = c.view
	} else {
		return value.ErrRef, nil
	}
	var (
		startLine = min(start.Line, end.Line)
		endLine   = max(start.Line, end.Line)
		startCol  = min(start.Column, end.Column)
		endCol    = max(start.Column, end.Column)
		height    = int(endLine - startLine + 1)
		width     = int(endCol - startCol + 1)
		data      = make([][]value.ScalarValue, height)
	)

	for i := 0; i < height; i++ {
		data[i] = make([]value.ScalarValue, width)

		for j := 0; j < width; j++ {
			pos := layout.Position{
				Line:   startLine + int64(i),
				Column: startCol + int64(j),
			}
			cell, err := sh.Cell(pos)
			if err != nil || cell == nil {
				data[i][j] = nil
				continue
			}
			data[i][j] = cell.Value()
		}
	}
	return value.NewArray(data), nil
}

func (c sheetContext) At(pos layout.Position) (value.Value, error) {
	if pos.Sheet == "" || pos.Sheet == c.view.Name() {
		cell, err := c.view.Cell(pos)
		if err != nil || cell == nil {
			return value.ErrRef, nil
		}
		if err := cell.Reload(c); err != nil && !errors.Is(err, ErrSupported) {
			return nil, err
		}
		return cell.Value(), nil
	}
	return value.ErrRef, nil
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

func (c fileContext) Resolve(name string) (value.Value, error) {
	return nil, ErrSupported
}

func (c fileContext) At(pos layout.Position) (value.Value, error) {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return nil, err
	}
	return SheetContext(sh).At(pos)
}

func (c fileContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return nil, fmt.Errorf("cross sheet range not allowed")
	}
	sh, err := c.sheet(start.Sheet)
	if err != nil {
		return value.ErrRef, nil
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
