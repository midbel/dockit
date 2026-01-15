package formula

import (
	"fmt"
	"math"

	"github.com/midbel/dockit/value"
)

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
		return nil, fmt.Errorf("unuspported expression type: %T", expr)
	}
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
	var args []value.Value
	for _, a := range e.args {
		v, err := Eval(a, ctx)
		if err != nil {
			return v, err
		}
		args = append(args, v)
	}
	fn, err := ctx.Resolve(id.name)
	if err != nil {
		return nil, err
	}
	if fn.Kind() != value.KindFunction {
		return nil, fmt.Errorf("%s is not callable", id)
	}
	call, ok := fn.(value.FunctionValue)
	return call.Call(args)

}

func evalCellAddr(e cellAddr, ctx value.Context) (value.Value, error) {
	return ctx.At(e.Position)
}

func evalRangeAddr(e rangeAddr, ctx value.Context) (value.Value, error) {
	return ctx.Range(e.startAddr.Position, e.endAddr.Position)
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
	ls := left.(value.ScalarValue)
	rs := right.(value.ScalarValue)
	res, err := do(ls.Scalar().(float64), rs.Scalar().(float64))
	if err != nil {
		return nil, err
	}
	return Float(res), nil
}
