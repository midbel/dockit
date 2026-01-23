package formula

import (
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/midbel/dockit/value"
)

var (
	ErrEval     = errors.New("expression can not be evaluated")
	ErrCallable = errors.New("expression is not callable")
)

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

func Exec(r io.Reader, env *Environment) (value.Value, error) {
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
		fmt.Printf("%+v\n", expr)
		if err != nil {
			return nil, err
		}
		if phase, err = execPhase(expr, phase); err != nil {
			return nil, err
		}
		if val, err = exec(expr, env); err != nil {
			return nil, err
		}
	}
	return val, nil
}

func exec(expr Expr, ctx *Environment) (value.Value, error) {
	switch e := expr.(type) {
	case useFile:
	case importFile:
	case printRef:
	case access:
	case literal:
		return Text(e.value), nil
	case number:
		return Float(e.value), nil
	default:
		return nil, ErrEval
	}
	return nil, nil
}

func Eval(expr Expr, ctx value.Context) (value.Value, error) {
	switch e := expr.(type) {
	case binary:
		return evalBinary(e, ctx)
	case unary:
		return evalUnary(e, ctx)
	case literal:
		return Text(e.value), nil
	case number:
		return Float(e.value), nil
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

func evalImport(e importFile, ctx value.Context) (value.Value, error) {
	return nil, nil
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

	if !IsScalar(left) && IsScalar(right) {
		return ErrValue, nil
	}

	switch e.op {
	case Add:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return left + right, nil
		})
	case Sub:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return left - right, nil
		})
	case Mul:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return left * right, nil
		})
	case Div:
		return doMath(left, right, func(left, right float64) (float64, error) {
			if right == 0 {
				return 0, ErrDiv0
			}
			return left / right, nil
		})
	case Pow:
		return doMath(left, right, func(left, right float64) (float64, error) {
			return math.Pow(left, right), nil
		})
	case Concat:
		if !IsScalar(left) || !IsScalar(right) {
			return ErrValue, nil
		}
		return Text(left.String() + right.String()), nil
	case Eq:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			return left.Equal(right)
		})
	case Ne:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			ok, err := left.Equal(right)
			return !ok, err
		})
	case Lt:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			return left.Less(right)
		})
	case Le:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			if ok, err := left.Equal(right); err == nil && ok {
				return ok, nil
			}
			return left.Less(right)
		})
	case Gt:
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
	case Ge:
		return doCmp(left, right, func(left value.Comparable, right value.Value) (bool, error) {
			if ok, err := left.Equal(right); err == nil && ok {
				return ok, nil
			}
			ok, err := left.Less(right)
			return !ok, err
		})
	default:
		return ErrValue, nil
	}
}

func evalUnary(e unary, ctx value.Context) (value.Value, error) {
	val, err := Eval(e.right, ctx)
	if err != nil {
		return nil, err
	}
	n, ok := val.(Float)
	if !ok {
		return ErrValue, nil
	}
	switch e.op {
	case Add:
		return n, nil
	case Sub:
		return Float(float64(-n)), nil
	default:
		return ErrValue, nil
	}
}

func evalCall(e call, ctx value.Context) (value.Value, error) {
	id, ok := e.ident.(identifier)
	if !ok {
		return ErrName, nil
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

func IsComparable(v value.Value) bool {
	_, ok := v.(value.Comparable)
	return ok
}

func IsNumber(v value.Value) bool {
	_, ok := v.(Float)
	return ok
}

func IsScalar(v value.Value) bool {
	return v.Kind() == value.KindScalar
}

func doMath(left, right value.Value, do func(left, right float64) (float64, error)) (value.Value, error) {
	if !IsNumber(left) {
		return ErrValue, nil
	}
	if !IsNumber(right) {
		return ErrValue, nil
	}
	var (
		ls = left.(value.ScalarValue)
		rs = right.(value.ScalarValue)
	)
	res, err := do(ls.Scalar().(float64), rs.Scalar().(float64))
	if err != nil {
		return nil, err
	}
	return Float(res), nil
}

func doCmp(left, right value.Value, do func(left value.Comparable, right value.Value) (bool, error)) (value.Value, error) {
	if !IsComparable(left) {
		return ErrValue, nil
	}
	ok, err := do(left.(value.Comparable), right)
	if err != nil {
		return ErrValue, nil
	}
	return Boolean(ok), nil
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
	src.Predicate, err = createPredicate(b.op, v)
	if err != nil {
		return nil, false, err
	}
	src.Value, err = Eval(b.left, ctx)
	if err != nil {
		return nil, false, err
	}
	return &src, true, err
}
