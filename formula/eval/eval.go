package eval

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

var (
	ErrEval     = errors.New("expression can not be evaluated")
	ErrCallable = errors.New("expression is not callable")
)

func Eval(expr Expr, ctx value.Context) (value.Value, error) {
	switch e := expr.(type) {
	case binary:
		return evalBinary(e, ctx)
	case unary:
		return evalUnary(e, ctx)
	case literal:
		return types.Text(e.value), nil
	case number:
		return types.Float(e.value), nil
	case call:
		return evalCall(e, ctx)
	case cellAddr:
		return evalCellAddr(e, ctx)
	case rangeAddr:
		return evalRangeAddr(e, ctx)
	default:
		return nil, ErrEval
	}
}

type scriptPhase int8

const (
	phaseStmt scriptPhase = 1 << iota
	phaseUse
	phaseImport
)

func (p scriptPhase) Allows(k Kind) bool {
	switch p {
	case phaseStmt:
		return k == KindStmt
	case phaseUse:
		return k == KindStmt || k == KindUse || k == KindImport
	case phaseImport:
		return k == KindStmt || k == KindImport
	default:
		return false
	}
}

func (p scriptPhase) Next(k Kind) scriptPhase {
	switch {
	case p == phaseUse && k == KindUse:
		return p
	case p == phaseUse && k == KindImport:
		return phaseImport
	case p == phaseUse && k == KindStmt:
		return phaseStmt
	case p == phaseImport && k == KindImport:
		return p
	case p == phaseImport && k == KindStmt:
		return phaseStmt
	case p == phaseStmt && k == KindStmt:
		return p
	default:
		return p
	}
}

func execPhase(expr Expr, phase scriptPhase) (scriptPhase, error) {
	currKind := KindStmt
	if ek, ok := expr.(ExprKind); ok {
		currKind = ek.Kind()
	}
	if !phase.Allows(currKind) {
		return phase, fmt.Errorf("unknown script phase!")
	}
	return phase.Next(currKind), nil
}

type Loader interface {
	Open(string) (grid.File, error)
}

type noopLoader struct{}

func (noopLoader) Open(_ string) (grid.File, error) {
	return nil, fmt.Errorf("noop loader can not open file")
}

func defaultLoader() Loader {
	return noopLoader{}
}

type Engine struct {
	Loader
	Stdout io.Writer
	Stderr io.Writer
}

func NewEngine(loader Loader) *Engine {
	e := Engine{
		Loader: loader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return &e
}

func (e *Engine) Exec(r io.Reader, ctx *env.Environment) (value.Value, error) {
	var (
		val   value.Value
		phase = phaseUse
		ps    = NewParser(ScriptGrammar())
	)
	if err := ps.Init(r); err != nil {
		return nil, err
	}
	for {
		expr, err := ps.ParseNext()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if phase, err = execPhase(expr, phase); err != nil {
			return nil, err
		}
		if val, err = e.exec(expr, ctx); err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (e *Engine) exec(expr Expr, ctx *env.Environment) (value.Value, error) {
	switch expr := expr.(type) {
	case importFile:
		return evalImport(e, expr, ctx)
	case useFile:
	case printRef:
		return evalPrint(e, expr, ctx)
	case access:
		return evalAccess(e, expr, ctx)
	case literal:
		return types.Text(expr.value), nil
	case number:
		return types.Float(expr.value), nil
	case identifier:
		return ctx.Resolve(expr.name)
	default:
		return nil, ErrEval
	}
	return nil, nil
}

func evalImport(eg *Engine, e importFile, ctx *env.Environment) (value.Value, error) {
	file, err := eg.Loader.Open(e.file)
	if err != nil {
		return nil, err
	}
	if e.alias == "" {
		var (
			alias string
			file  = e.file
		)
		for {
			ext := filepath.Ext(file)
			if ext == "" {
				break
			}
			alias = strings.TrimSuffix(file, ext)
		}
		e.alias = alias
	}
	ctx.Define(e.alias, types.NewFileValue(file))
	return types.Empty(), nil
}

func evalPrint(eg *Engine, e printRef, ctx *env.Environment) (value.Value, error) {
	v, err := eg.exec(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(eg.Stdout, v.String())
	return types.Empty(), nil
}

func evalAccess(eg *Engine, e access, ctx *env.Environment) (value.Value, error) {
	obj, err := eg.exec(e.expr, ctx)
	if err != nil {
		return nil, err
	}
	g, ok := obj.(value.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("object expected")
	}
	return g.Get(e.prop)
}

func evalBinary(e binary, ctx value.Context) (value.Value, error) {
	left, err := Eval(e.left, ctx)
	if err != nil {
		return nil, err
	}
	right, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}

	if !types.IsScalar(left) && types.IsScalar(right) {
		return types.ErrValue, nil
	}

	switch e.op {
	case op.Add:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return left + right, nil
		})
	case op.Sub:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return left - right, nil
		})
	case op.Mul:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return left * right, nil
		})
	case op.Div:
		return doMath(left, right, func(left, right float64) (float64, error) {
			if right == 0 {
				return 0, types.ErrDiv0
			}
			return left / right, nil
		})
	case op.Pow:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return math.Pow(left, right), nil
		})
	case op.Concat:
		if !types.IsScalar(left) || !types.IsScalar(right) {
			return types.ErrValue, nil
		}
		return types.Text(left.String() + right.String()), nil
	case op.Eq:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			return left.Equal(right)
		})
	case op.Ne:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			ok, err := left.Equal(right)
			return !ok, err
		})
	case op.Lt:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			return left.Less(right)
		})
	case op.Le:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			if ok, err := left.Equal(right); err == nil && ok {
				return ok, nil
			}
			return left.Less(right)
		})
	case op.Gt:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			if ok, err := left.Equal(right); err == nil && ok {
				return !ok, nil
			}
			ok, err := left.Less(right)
			if !ok && err == nil {
				ok = !ok
			}
			return ok, err
		})
	case op.Ge:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			if ok, err := left.Equal(right); err == nil && ok {
				return ok, nil
			}
			ok, err := left.Less(right)
			return !ok, err
		})
	default:
		return types.ErrValue, nil
	}
}

func evalUnary(e unary, ctx value.Context) (value.Value, error) {
	val, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(types.Float)
	if !ok {
		return types.ErrValue, nil
	}
	switch e.op {
	case op.Add:
		return n, nil
	case op.Sub:
		return types.Float(float64(-n)), nil
	default:
		return types.ErrValue, nil
	}
}

func evalCall(e call, ctx value.Context) (value.Value, error) {
	id, ok := e.ident.(identifier)
	if !ok {
		return types.ErrName, nil
	}
	var args []value.Arg
	for i := range e.args {
		args = append(args, makeArg(e.args[i]))
	}
	fn, err := ctx.Resolve(id.name)
	if err != nil {
		return nil, err
	}
	if fn.Kind() != value.KindFunction {
		return nil, fmt.Errorf("%s: %w", id.name, ErrCallable)
	}
	call, ok := fn.(value.FunctionValue)
	return call.Call(args, ctx)
}

func evalCellAddr(e cellAddr, ctx value.Context) (value.Value, error) {
	return ctx.At(e.Position)
}

func evalRangeAddr(e rangeAddr, ctx value.Context) (value.Value, error) {
	return ctx.Range(e.startAddr.Position, e.endAddr.Position)
}

func doMath(left, right value.Value, do func(left, right float64) (float64, error)) (value.Value, error) {
	if !types.IsNumber(left) {
		return types.ErrValue, nil
	}
	if !types.IsNumber(right) {
		return types.ErrValue, nil
	}
	var (
		ls = left.(value.ScalarValue)
		rs = right.(value.ScalarValue)
	)
	res, err := do(ls.Scalar().(float64), rs.Scalar().(float64))
	if err != nil {
		return nil, err
	}
	return types.Float(res), nil
}

func doCmp(left, right value.Value, do func(left value.Comparable, right value.Value) (bool, error)) (value.Value, error) {
	if !types.IsComparable(left) {
		return types.ErrValue, nil
	}
	ok, err := do(left.(value.Comparable), right)
	if err != nil {
		return types.ErrValue, nil
	}
	return types.Boolean(ok), nil
}

type arg struct {
	expr Expr
}

func makeArg(expr Expr) value.Arg {
	return arg{
		expr: expr,
	}
}

func (a arg) Eval(ctx value.Context) (value.Value, error) {
	return Eval(a.expr, ctx)
}

func (a arg) asFilter(ctx value.Context) (*value.Filter, bool, error) {
	var src value.Filter

	b, ok := a.expr.(binary)
	if !ok {
		return nil, false, fmt.Errorf("argument can not be used as a predicate")
	}
	v, err := Eval(b.right, ctx)
	if err != nil {
		return nil, false, err
	}
	src.Predicate, err = types.NewPredicate(b.op, v)
	if err != nil {
		return nil, false, err
	}
	src.Value, err = Eval(b.left, ctx)
	if err != nil {
		return nil, false, err
	}
	return &src, true, err
}
