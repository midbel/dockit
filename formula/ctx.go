package formula

import (
	"errors"
	"fmt"
	"os"

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
