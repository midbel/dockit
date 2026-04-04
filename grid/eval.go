package grid

import (
	"fmt"
	"iter"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func Eval(expr value.Formula, ctx value.Context) (value.Value, error) {
	e, ok := expr.(formula)
	if !ok {
		return value.ErrValue, fmt.Errorf("formula can not eval")
	}
	return eval(e.expr, ctx), nil
}

func EvalString(expr string, ctx value.Context) (value.Value, error) {
	e, err := ParseFormula(expr)
	if err != nil {
		return nil, err
	}
	return Eval(e, ctx)
}

func eval(expr parse.Expr, ctx value.Context) value.Value {
	switch e := expr.(type) {
	case parse.Binary:
		return evalBinary(e, ctx)
	case parse.Unary:
		return evalUnary(e, ctx)
	case parse.Postfix:
		return evalPostfix(e, ctx)
	case parse.Identifier:
		return evalIdentifier(e, ctx)
	case parse.Literal:
		return value.Text(e.Text())
	case parse.Number:
		return value.Float(e.Float())
	case parse.Call:
		return evalCall(e, ctx)
	case parse.CellAddr:
		return evalCellAddr(e, ctx)
	case parse.RangeAddr:
		return evalRangeAddr(e, ctx)
	default:
		return value.ErrNA
	}
}

func evalIdentifier(e parse.Identifier, ctx value.Context) value.Value {
	switch e.Ident() {
	case "true":
		return value.Boolean(true)
	case "false":
		return value.Boolean(false)
	case value.ErrNull.String():
		return value.ErrNull
	case value.ErrDiv0.String():
		return value.ErrDiv0
	case value.ErrValue.String():
		return value.ErrValue
	case value.ErrRef.String():
		return value.ErrRef
	case value.ErrName.String():
		return value.ErrName
	case value.ErrNum.String():
		return value.ErrNum
	case value.ErrNA.String():
		return value.ErrNA
	default:
		col, _ := layout.ParseIndex(e.Ident())
		return ctx.At(layout.NewPosition(0, col))
	}
}

func evalBinary(e parse.Binary, ctx value.Context) value.Value {
	var (
		left  = eval(e.Left(), ctx)
		right = eval(e.Right(), ctx)
	)
	if err := value.HasErrors(left, right); err != nil {
		return err
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
		return value.Boolean(ok)
	case op.Or:
		ok := value.True(left) || value.True(right)
		return value.Boolean(ok)
	default:
		return value.ErrValue
	}
}

func evalPostfix(e parse.Postfix, ctx value.Context) value.Value {
	val := eval(e.Expr(), ctx)
	if value.IsError(val) {
		return val
	}
	switch e.Op() {
	case op.Percent:
		return value.Div(val, value.Float(100))
	default:
		return value.ErrValue
	}
}

func evalUnary(e parse.Unary, ctx value.Context) value.Value {
	val := eval(e.Expr(), ctx)
	if value.IsError(val) {
		return val
	}
	n, err := value.CastToFloat(val)
	if err != nil {
		return value.ErrValue
	}
	switch e.Op() {
	case op.Not:
		ok := value.True(val)
		return value.Boolean(!ok)
	case op.Add:
		return n
	case op.Sub:
		return value.Float(float64(-n))
	default:
		return value.ErrValue
	}
}

func evalCall(e parse.Call, ctx value.Context) value.Value {
	id, ok := e.Name().(parse.Identifier)
	if !ok {
		return value.ErrName
	}
	var args []value.Value
	for _, e := range e.Args() {
		args = append(args, newArg(e, ctx))
	}
	fn, err := builtins.Lookup(id.Ident())
	if err != nil {
		return value.ErrName
	}
	return fn(args)
}

func evalCellAddr(e parse.CellAddr, ctx value.Context) value.Value {
	val := ctx.At(e.Position)
	if f, ok := val.(value.Formula); ok {
		v, err := Eval(f, ctx)
		if err != nil {
			return value.ErrValue
		}
		return v
	}
	return val
}

func evalRangeAddr(e parse.RangeAddr, ctx value.Context) value.Value {
	return ctx.Range(e.StartAt().Position, e.EndAt().Position)
}

func ParseFormula(str string) (value.Formula, error) {
	expr, err := parse.ParseFormula(str)
	if err != nil {
		return nil, err
	}
	return NewFormula(expr), nil
}

type arg struct {
	expr parse.Expr
	ctx  value.Context
}

func newArg(expr parse.Expr, ctx value.Context) value.Value {
	return arg{
		expr: expr,
		ctx:  ctx,
	}
}

func (arg) Type() string {
	return "argument"
}

func (arg) Kind() value.ValueKind {
	return value.KindScalar
}

func (a arg) String() string {
	return a.expr.String()
}

func (a arg) Eval() value.Value {
	return eval(a.expr, a.ctx)
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

func (f formula) Eval(ctx value.Context) value.Value {
	return eval(f.expr, ctx)
}

func (f formula) Clone(pos layout.Position) value.Formula {
	if c, ok := f.expr.(parse.Clonable); ok {
		return NewFormula(c.CloneWithOffset(pos))
	}
	return f
}

type arrayView struct {
	inner View
}

func ArrayView(view View) value.Value {
	v := arrayView{
		inner: view,
	}
	return &v
}

func (*arrayView) Type() string {
	return value.TypeArray
}

func (*arrayView) Kind() value.ValueKind {
	return value.KindArray
}

func (v *arrayView) String() string {
	return fmt.Sprintf("%s(%s)", value.TypeArray, v.inner.Name())
}

func (v *arrayView) Dimension() layout.Dimension {
	rg := v.inner.Bounds()
	dm := layout.Dimension{
		Lines:   int64(rg.Height()),
		Columns: int64(rg.Width()),
	}
	return dm
}

func (v *arrayView) At(row, col int) value.ScalarValue {
	pos := layout.Position{
		Line:   int64(row) + 1,
		Column: int64(col) + 1,
	}
	c, _ := v.inner.Cell(pos)
	return c.Value()
}

func (a arrayView) Values() iter.Seq[value.ScalarValue] {
	it := func(yield func(value.ScalarValue) bool) {
		for _, rs := range a.inner.Rows() {
			for _, v := range rs {
				ok := yield(v)
				if !ok {
					return
				}
			}
		}
	}
	return it
}
