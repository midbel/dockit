package grid

import (
	"fmt"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/value"
	"github.com/midbel/dockit/layout"
)

func Eval(expr value.Formula, ctx value.Context) (value.Value, error) {
	e, ok := expr.(formula)
	if !ok {
		return nil, fmt.Errorf("formula can not eval")
	}
	return eval(e.expr, ctx)
}

func eval(expr parse.Expr, ctx value.Context) (value.Value, error) {
	switch e := expr.(type) {
	case parse.Binary:
		return evalBinary(e, ctx)
	case parse.Unary:
		return evalUnary(e, ctx)
	case parse.Literal:
		return value.Text(e.Text()), nil
	case parse.Number:
		return value.Float(e.Float()), nil
	case parse.Call:
		return evalCall(e, ctx)
	case parse.CellAddr:
		return evalCellAddr(e, ctx)
	case parse.RangeAddr:
		return evalRangeAddr(e, ctx)
	default:
		return value.ErrValue, nil
	}	
}

func evalBinary(e parse.Binary, ctx value.Context) (value.Value, error) {
	left, err := eval(e.Left(), ctx)
	if err != nil {
		return nil, err
	}
	right, err := eval(e.Right(), ctx)
	if err != nil {
		return nil, err
	}

	switch e.Op() {
	case op.Add:
		return value.Add(left, right)
	case op.Sub:
		return value.Sub(left, right)
	case op.Mul:
		return value.Mul(left, right)
	case op.Div:
		return value.Div(left, right)
	case op.Pow:
		return value.Pow(left, right)
	case op.Concat:
		return value.Concat(left, right)
	case op.Eq:
		return value.Eq(left, right)
	case op.Ne:
		return value.Ne(left, right)
	case op.Lt:
		return value.Lt(left, right)
	case op.Le:
		return value.Le(left, right)
	case op.Gt:
		return value.Gt(left, right)
	case op.Ge:
		return value.Ge(left, right)
	case op.And:
		ok := value.True(left) && value.True(right)
		return value.Boolean(ok), nil
	case op.Or:
		ok := value.True(left) || value.True(right)
		return value.Boolean(ok), nil
	default:
		return value.ErrValue, nil
	}
}

func evalUnary(e parse.Unary, ctx value.Context) (value.Value, error) {
	val, err := eval(e.Expr(), ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(value.Float)
	if !ok {
		return value.ErrValue, nil
	}
	switch e.Op() {
	case op.Not:
		ok := value.True(val)
		return value.Boolean(!ok), nil
	case op.Add:
		return n, nil
	case op.Sub:
		return value.Float(float64(-n)), nil
	default:
		return value.ErrValue, nil
	}
}

func evalCall(e parse.Call, ctx value.Context) (value.Value, error) {
	id, ok := e.Name().(parse.Identifier)
	if !ok {
		return value.ErrName, nil
	}
	var args []value.Arg
	for _, a := range e.Args() {
		args = append(args, makeArg(a))
	}
	fn, err := ctx.Resolve(id.Ident())
	if err != nil {
		return nil, err
	}
	if fn.Kind() != value.KindFunction {
		return nil, fmt.Errorf("%s: not a function", id.Ident())
	}
	call, ok := fn.(value.FunctionValue)
	return call.Call(args, ctx)
}

func evalCellAddr(e parse.CellAddr, ctx value.Context) (value.Value, error) {
	return ctx.At(e.Position)
}

func evalRangeAddr(e parse.RangeAddr, ctx value.Context) (value.Value, error) {
	return ctx.Range(e.StartAt().Position, e.EndAt().Position)
}

func ParseFormula(str string) (value.Formula, error) {
	expr, err := parse.ParseFormula(str)
	if err != nil {
		return nil, err
	}
	return NewFormula(expr), nil
}

type formula struct {
	expr parse.Expr
}

func NewFormula(expr parse.Expr) value.Formula {
	return formula{
		expr: expr,
	}
}

func (formula) Type() string {
	return "formula"
}

func (formula) Kind() value.ValueKind {
	return value.KindFunction
}

func (f formula) String() string {
	return f.expr.String()
}

func (f formula) Eval(ctx value.Context) (value.Value, error) {
	return eval(f.expr, ctx)
}

func (f formula) Clone(pos layout.Position) value.Formula {
	if c, ok := f.expr.(parse.Clonable); ok {
		return NewFormula(c.CloneWithOffset(pos))
	}
	return f
}

type arg struct {
	expr parse.Expr
}

func makeArg(expr parse.Expr) value.Arg {
	return arg{
		expr: expr,
	}
}

func (a arg) Eval(ctx value.Context) (value.Value, error) {
	return eval(a.expr, ctx)
}