package oxml

import (
	"fmt"

	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type sheetContext struct {
	view   grid.View
	parent value.Context
}

func SheetContext(parent value.Context, sheet grid.View) value.Context {
	return sheetContext{
		parent: parent,
		view:   sheet,
	}
}

func (c sheetContext) Resolve(_ string) (value.Value, error) {
	return nil, grid.ErrSupported
}

func (c sheetContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return formula.ErrRef, nil
	}
	var sh grid.View
	if start.Sheet == "" || start.Sheet == c.view.Name() {
		sh = c.view
	} else {
		if c.parent == nil {
			return formula.ErrRef, nil
		}
		return c.parent.Range(start, end)
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

	arr := formula.Array{
		Data: data,
	}
	return arr, nil
}

func (c sheetContext) At(pos layout.Position) (value.Value, error) {
	if pos.Sheet == "" || pos.Sheet == c.view.Name() {
		cell, err := c.view.Cell(pos)
		if err != nil || cell == nil {
			return formula.ErrRef, nil
		}
		return cell.Value(), nil
	}
	if c.parent == nil {
		return formula.ErrRef, nil
	}
	return c.parent.At(pos)
}

func (c sheetContext) ResolveFunc(name string) (formula.BuiltinFunc, error) {
	fn, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("%s not defined", name)
	}
	return fn, nil
}

type fileContext struct {
	*File
}

func FileContext(file *File) value.Context {
	return fileContext{
		File: file,
	}
}

func (c fileContext) Resolve(_ string) (value.Value, error) {
	return nil, grid.ErrSupported
}

func (c fileContext) At(pos layout.Position) (value.Value, error) {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return formula.ErrRef, nil
	}
	ctx := SheetContext(nil, sh)
	return ctx.At(pos)
}

func (c fileContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return formula.ErrRef, nil
	}
	sh, err := c.sheet(start.Sheet)
	if err != nil {
		return formula.ErrRef, nil
	}
	ctx := SheetContext(nil, sh)
	return ctx.Range(start, end)
}

func (c fileContext) ResolveFunc(name string) (formula.BuiltinFunc, error) {
	fn, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("%s not defined", name)
	}
	return fn, nil
}

func (c fileContext) sheet(name string) (grid.View, error) {
	var (
		sh  grid.View
		err error
	)
	if name == "" {
		sh, err = c.File.ActiveSheet()
	} else {
		sh, err = c.File.Sheet(name)
	}
	return sh, err
}
