package grid

import (
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

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
		if pos.Line == 0 {
			return c.fromColumn(pos)
		} else if pos.Column == 0 {
			return c.fromRow(pos)
		}
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

func (c sheetContext) fromRow(pos layout.Position) value.Value {
	var (
		bd    = c.view.Bounds()
		start = layout.NewPosition(pos.Line, 1)
		end   = layout.NewPosition(pos.Line, bd.Width())
	)
	return c.Range(start, end)
}

func (c sheetContext) fromColumn(pos layout.Position) value.Value {
	var (
		bd    = c.view.Bounds()
		start = layout.NewPosition(1, pos.Column)
		end   = layout.NewPosition(bd.Height(), pos.Column)
	)
	return c.Range(start, end)
}

type fileContext struct {
	file File
}

func FileContext(file File) value.Context {
	return fileContext{
		file: file,
	}
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

func (c fileContext) sheet(name string) (View, error) {
	if name == "" {
		return c.file.ActiveSheet()
	}
	return c.file.Sheet(name)
}

type evalContext struct {
	parent value.Context
	child  value.Context
}

func EnclosedContext(parent, child value.Context) value.Context {
	return evalContext{
		parent: parent,
		child:  child,
	}
}

func (c evalContext) Resolve(ident string) value.Value {
	return value.ErrValue
}

func (c evalContext) At(pos layout.Position) value.Value {
	if c.child != nil {
		val := c.child.At(pos)
		if !value.IsError(val) {
			return val
		}
	}
	if c.parent != nil {
		return c.parent.At(pos)
	}
	return value.ErrValue
}

func (c evalContext) Range(start, end layout.Position) value.Value {
	if c.child != nil {
		val := c.child.Range(start, end)
		if !value.IsError(val) {
			return val
		}
	}
	if c.parent != nil {
		return c.parent.Range(start, end)
	}
	return value.ErrValue
}
