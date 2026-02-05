package eval

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrSupported = errors.New("not supported")
	ErrEmpty     = errors.New("empty context")
	ErrMutate    = errors.New("context is not mutable")
)

type scopedContext []value.Context

func EvalContext(others ...value.Context) value.Context {
	es := scopedContext(others)
	return &es
}

func (ec *scopedContext) Push(ctx value.Context) {
	*ec = append(*ec, ctx)
}

func (ec *scopedContext) Pop() {
	if n := len(*ec); n >= 1 {
		*ec = (*ec)[:n-1]
	}
}

func (ec *scopedContext) ReadOnly() value.Context {
	return value.ReadOnly(ec)
}

func (ec *scopedContext) Resolve(name string) (value.Value, error) {
	for i := len(*ec) - 1; i >= 0; i-- {
		v, err := (*ec)[i].Resolve(name)
		if err == nil {
			return v, err
		}
	}
	return types.ErrValue, nil
}

func (ec *scopedContext) At(pos layout.Position) (value.Value, error) {
	for i := len(*ec) - 1; i >= 0; i-- {
		v, err := (*ec)[i].At(pos)
		if err == nil {
			return v, err
		}
	}
	return types.ErrValue, nil
}

func (ec *scopedContext) Range(start, end layout.Position) (value.Value, error) {
	for i := len(*ec) - 1; i >= 0; i-- {
		v, err := (*ec)[i].Range(start, end)
		if err == nil {
			return v, err
		}
	}
	return types.ErrValue, nil
}

func (ec *scopedContext) SetValue(pos layout.Position, val value.Value) error {
	ctx := ec.top()
	if ctx == nil {
		return ErrEmpty
	}
	if mc, ok := ctx.(value.MutableContext); ok {
		return mc.SetValue(pos, val)
	}
	return nil
}

func (ec *scopedContext) SetFormula(pos layout.Position, val value.Formula) error {
	ctx := ec.top()
	if ctx == nil {
		return ErrEmpty
	}
	if mc, ok := ctx.(value.MutableContext); ok {
		return mc.SetFormula(pos, val)
	}
	return nil
}

func (ec *scopedContext) SetRange(start, end layout.Position, val value.Value) error {
	// TBD
	return nil
}

func (ec *scopedContext) SetRangeFormula(start, end layout.Position, val value.Value) error {
	// TBD
	return nil
}

func (ec *scopedContext) top() value.Context {
	n := len(*ec)
	if n > 0 {
		return (*ec)[n-1]
	}
	return nil
}

type sheetContext struct {
	view grid.View
}

func SheetContext(sheet grid.View) value.Context {
	return sheetContext{
		view: sheet,
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
	var sh grid.View
	if start.Sheet == "" || start.Sheet == c.view.Name() {
		sh = c.view
	} else {
		return types.ErrRef, nil
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
	return types.NewArray(data), nil
}

func (c sheetContext) At(pos layout.Position) (value.Value, error) {
	if pos.Sheet == "" || pos.Sheet == c.view.Name() {
		cell, err := c.view.Cell(pos)
		if err != nil || cell == nil {
			return types.ErrRef, nil
		}
		return cell.Value(), nil
	}
	return types.ErrRef, nil
}

func (c sheetContext) SetValue(pos layout.Position, val value.Value) error {
	mv, ok := c.view.(grid.MutableView)
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
	mv, ok := c.view.(grid.MutableView)
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
	file grid.File
}

func FileContext(file grid.File) value.Context {
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
	ctx := EvalContext(c, SheetContext(sh))
	return ctx.At(pos)
}

func (c fileContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return nil, fmt.Errorf("cross sheet range not allowed")
	}
	sh, err := c.sheet(start.Sheet)
	if err != nil {
		return types.ErrRef, nil
	}
	ctx := EvalContext(c, SheetContext(sh))
	return ctx.Range(start, end)
}

func (c fileContext) SetValue(pos layout.Position, val value.Value) error {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return err
	}
	mv, ok := sh.(grid.MutableView)
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
	mv, ok := sh.(grid.MutableView)
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

func (c fileContext) sheet(name string) (grid.View, error) {
	if name == "" {
		return c.file.ActiveSheet()
	} else {
	}
	return c.file.Sheet(name)
}
