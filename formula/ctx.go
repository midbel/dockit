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

type fileValue struct {
	file grid.File
}

func NewFileValue(file grid.File) value.Value {
	return &fileValue{
		file: file,
	}
}

func (*fileValue) Kind() value.ValueKind {
	return value.KindObject
}

func (*fileValue) String() string {
	return "workbook"
}

func (c *fileValue) Get(ident string) (value.Value, error) {
	switch ident {
	case "sheets":
		x := c.file.Sheets()
		return Float(float64(len(x))), nil
	case "protected":
		return Boolean(false), nil
	case "active":
		sh, err := c.file.ActiveSheet()
		if err != nil {
			return ErrValue, nil
		}
		return NewViewValue(sh), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
}

type viewValue struct {
	view View
}

func NewViewValue(view View) value.Value {
	return &viewValue{
		view: view,
	}
}

func (*viewValue) Kind() value.ValueKind {
	return value.KindObject
}

func (c *viewValue) String() string {
	return c.view.Name()
}

func (c *viewValue) Get(ident string) (value.Value, error) {
	switch ident {
	case "name":
		return Text(c.view.Name()), nil
	case "lines":
		rg := c.view.Bounds()
		lines := rg.Ends.Line - rg.Starts.Line
		return Float(float64(lines)), nil
	case "columns":
		rg := c.view.Bounds()
		lines := rg.Ends.Column - rg.Starts.Column
		return Float(float64(lines)), nil
	case "cells":
		var count int
		for x := range c.view.Rows() {
			count += len(x)
		}
		return Float(float64(count)), nil
	case "empty":
		return Float(float64(0)), nil
	case "protected":
		var locked bool
		if k, ok := c.view.(interface{ IsLock() bool }); ok {
			locked = k.IsLock()
		}
		return Boolean(locked), nil
	case "readonly":
		return Boolean(false), nil
	case "active":
		return Boolean(false), nil
	case "index":
		return Float(float64(0)), nil
	default:
		return nil, fmt.Errorf("%s: %w", ident, ErrUndefined)
	}
}

type rangeValue struct {
	rg *layout.Range
}

func (*rangeValue) Kind() value.ValueKind {
	return value.KindObject
}

func (v *rangeValue) String() string {
	return v.rg.String()
}

func (v *rangeValue) Get(name string) (value.ScalarValue, error) {
	return nil, nil
}

type lambdaValue struct {
	expr Expr
}

func (*lambdaValue) Kind() value.ValueKind {
	return value.KindFunction
}

func (*lambdaValue) String() string {
	return "<formula>"
}

func (v *lambdaValue) Call(args []value.Arg, ctx value.Context) (value.Value, error) {
	return Eval(v.expr, ctx)
}

type envValue struct{}

func (envValue) Kind() value.ValueKind {
	return value.KindObject
}

func (envValue) String() string {
	return "env"
}

func (v envValue) Get(name string) (value.ScalarValue, error) {
	str := os.Getenv(name)
	return Text(str), nil
}

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
