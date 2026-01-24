package formula

import (
	"errors"
	"fmt"
	"os"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

var (
	ErrUndefined = errors.New("undefined identifier")
	ErrAvailable = errors.New("not available")
)

type Builtin interface {
	Call(args []value.Value) (value.Value, error)
	Arity() int
	Variadic() bool
}

type ReducerFunc func(value.Predicate, value.Value) (value.Value, error)

type BuiltinFunc func([]value.Value) (value.Value, error)

type Environment struct {
	values map[string]value.Value
	parent value.Context
}

func Enclosed(parent value.Context) *Environment {
	ctx := Environment{
		values: make(map[string]value.Value),
		parent: parent,
	}
	return &ctx
}

func Empty() *Environment {
	return Enclosed(nil)
}

func (c *Environment) Resolve(ident string) (value.Value, error) {
	v, ok := c.values[ident]
	if ok {
		return v, nil
	}
	if c.parent == nil {
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
	return c.parent.Resolve(ident)
}

func (c *Environment) Define(ident string, val value.Value) {
	c.values[ident] = val
}

func (c *Environment) At(_ layout.Position) (value.Value, error) {
	return nil, ErrAvailable
}

func (c *Environment) Range(_, _ layout.Position) (value.Value, error) {
	return nil, ErrAvailable
}

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
		return nil, err, nil
	}
	var sh View
	if start.Sheet == "" || start.Sheet == c.view.Name() {
		sh = c.view
	} else {
		if c.parent == nil {
			return nil, err, nil
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
			return nil, err, nil
		}
		return cell.Value(), nil
	}
	if c.parent == nil {
		return nil, err, nil
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
		return nil, err, nil
	}
	ctx := SheetContext(c, sh)
	return ctx.At(pos)
}

func (c fileContext) Range(start, end layout.Position) (value.Value, error) {
	if start.Sheet != end.Sheet {
		return nil, err, nil
	}
	sh, err := c.sheet(start.Sheet)
	if err != nil {
		return nil, err, nil
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

type truePredicate struct{}

func (truePredicate) Test(value.ScalarValue) (bool, error) {
	return true, nil
}

type cmpPredicate struct {
	op     rune
	scalar value.ScalarValue
}

func (p cmpPredicate) Test(other value.ScalarValue) (bool, error) {
	c, ok := p.scalar.(value.Comparable)
	if !ok {
		return false, fmt.Errorf("value is not comparable")
	}
	var err error
	switch p.op {
	case Eq:
		ok, err = c.Equal(other)
	case Ne:
		ok, err = c.Equal(other)
		ok = !ok
	case Lt:
		return c.Less(other)
	case Le:
		ok, err = c.Equal(other)
		if ok && err == nil {
			break
		}
		ok, err = c.Less(other)
	case Gt:
	case Ge:
		ok, err = c.Equal(other)
		if ok && err == nil {
			break
		}
	default:
	}
	return ok, err
}

func createPredicate(op rune, val value.Value) (value.Predicate, error) {
	scalar, ok := val.(value.ScalarValue)
	if !ok {
		return nil, fmt.Errorf("predicate can only operate on scalar value")
	}
	var p value.Predicate
	switch op {
	case Eq, Ne, Lt, Le, Gt, Ge:
		p = cmpPredicate{
			op:     op,
			scalar: scalar,
		}
	default:
		return nil, fmt.Errorf("unsupported predicate type")
	}
	return p, nil
}

func callAny(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return Boolean(false), nil
}

func callAll(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return Boolean(false), nil
}

func callCount(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return Float(0), nil
}
