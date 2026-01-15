package grid

import (
	"fmt"

	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type viewValue struct {
	view View
}

func (*viewValue) Kind() value.ValueKind {
	return value.KindObject
}

func (c *viewValue) String() string {
	return c.view.Name()
}

func (c *viewValue) Get(ident string) (value.ScalarValue, error) {
	switch ident {
	case "name":
		return formula.Text(c.view.Name()), nil
	case "lines":
		rg := c.view.Bounds()
		lines := rg.Ends.Line - rg.Starts.Line
		return formula.Float(float64(lines)), nil
	case "columns":
		rg := c.view.Bounds()
		lines := rg.Ends.Column - rg.Starts.Column
		return formula.Float(float64(lines)), nil
	case "cells":
		var count int
		for x := range c.view.Rows() {
			count += len(x)
		}
		return formula.Float(float64(count)), nil
	case "empty":
		return formula.Float(float64(0)), nil
	case "protected":
		var locked bool
		if k, ok := c.view.(interface{ IsLock() bool }); ok {
			locked = k.IsLock()
		}
		return formula.Boolean(locked), nil
	case "readonly":
		return formula.Boolean(false), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, formula.ErrUndefined)
	}
}

type fileValue struct {
	file File
}

func (*fileValue) Kind() value.ValueKind {
	return value.KindObject
}

func (*fileValue) String() string {
	return "workbook"
}

func (c *fileValue) Get(ident string) (value.ScalarValue, error) {
	switch ident {
	case "sheets":
		x := c.file.Sheets()
		return formula.Float(float64(len(x))), nil
	case "protected":
		return formula.Boolean(false), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, formula.ErrUndefined)
	}
}

type sheetContext struct {
	view   View
	parent value.Context
}

func SheetContext(parent value.Context, sheet View) value.Context {
	return sheetContext{
		parent: parent,
		view:   sheet,
	}
}

func (c sheetContext) Resolve(name string) (value.Value, error) {
	if c.parent != nil {
		return c.parent.Resolve(name)
	}
	return nil, ErrSupported
}

func (c sheetContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return formula.ErrRef, nil
	}
	var sh View
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

type fileContext struct {
	file   File
	parent value.Context
}

func FileContext(parent value.Context, file File) value.Context {
	return fileContext{
		file:   file,
		parent: parent,
	}
}

func (c fileContext) Resolve(name string) (value.Value, error) {
	if c.parent != nil {
		return c.parent.Resolve(name)
	}
	return nil, ErrSupported
}

func (c fileContext) At(pos layout.Position) (value.Value, error) {
	sh, err := c.sheet(pos.Sheet)
	if err != nil {
		return formula.ErrRef, nil
	}
	ctx := SheetContext(c, sh)
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
	ctx := SheetContext(c, sh)
	return ctx.Range(start, end)
}

func (c fileContext) sheet(name string) (View, error) {
	if name == "" {
		return c.file.ActiveSheet()
	} else {
	}
	return c.file.Sheet(name)
}
