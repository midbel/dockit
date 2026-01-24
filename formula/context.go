package formula

import (
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
	"github.com/midbel/dockit/formula/types"
)

type sheetContext struct {
	view   grid.View
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
		return nil, fmt.Errorf("cross sheet range not allowed")
	}
	var sh View
	if start.Sheet == "" || start.Sheet == c.view.Name() {
		sh = c.view
	} else {
		if c.parent == nil {
			return nil, ErrUndefined
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

	arr := types.Array{
		Data: data,
	}
	return arr, nil
}

func (c sheetContext) At(pos layout.Position) (value.Value, error) {
	if pos.Sheet == "" || pos.Sheet == c.view.Name() {
		cell, err := c.view.Cell(pos)
		if err != nil || cell == nil {
			return nil, err
		}
		return cell.Value(), nil
	}
	if c.parent == nil {
		return nil, err
	}
	return c.parent.At(pos)
}

type fileContext struct {
	file   grid.File
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
		return nil, err
	}
	ctx := SheetContext(c, sh)
	return ctx.At(pos)
}

func (c fileContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return nil, err
	}
	sh, err := c.sheet(start.Sheet)
	if err != nil {
		return nil, err
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
