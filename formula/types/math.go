package types

import (
	"math"

	"github.com/midbel/dockit/value"
)

func Add(left, right value.Value) (value.Value, error) {
	ls, err := CastToFloat(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToFloat(right)
	if err != nil {
		return nil, err
	}
	return Float(ls + rs), nil
}

func Sub(left, right value.Value) (value.Value, error) {
	ls, err := CastToFloat(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToFloat(right)
	if err != nil {
		return nil, err
	}
	return Float(ls - rs), nil
}

func Mul(left, right value.Value) (value.Value, error) {
	ls, err := CastToFloat(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToFloat(right)
	if err != nil {
		return nil, err
	}
	return Float(ls * rs), nil
}

func Div(left, right value.Value) (value.Value, error) {
	ls, err := CastToFloat(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToFloat(right)
	if err != nil {
		return nil, err
	}
	if rs == 0 {
		return ErrDiv0, nil
	}
	return Float(ls / rs), nil
}

func Pow(left, right value.Value) (value.Value, error) {
	ls, err := CastToFloat(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToFloat(right)
	if err != nil {
		return nil, err
	}
	return Float(math.Pow(float64(ls), float64(rs))), nil
}

func Concat(left, right value.Value) (value.Value, error) {
	ls, err := CastToText(left)
	if err != nil {
		return nil, err
	}
	rs, err := CastToText(right)
	if err != nil {
		return nil, err
	}
	return Text(ls + rs), nil
}

func Eq(left, right value.Value) (value.Value, error) {
	cmp, ok := left.(value.Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	return Boolean(ok), err
}

func Ne(left, right value.Value) (value.Value, error) {
	cmp, ok := left.(value.Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	return Boolean(!ok), err
}

func Lt(left, right value.Value) (value.Value, error) {
	cmp, ok := left.(value.Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Less(right)
	return Boolean(!ok), err
}

func Le(left, right value.Value) (value.Value, error) {
	cmp, ok := left.(value.Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(ok), nil
	}
	ok, err = cmp.Less(right)
	return Boolean(ok), err
}

func Gt(left, right value.Value) (value.Value, error) {
	cmp, ok := left.(value.Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(!ok), nil
	}
	ok, err = cmp.Less(right)
	return Boolean(!ok), err
}

func Ge(left, right value.Value) (value.Value, error) {
	cmp, ok := left.(value.Comparable)
	if !ok {
		return ErrValue, nil
	}
	ok, err := cmp.Equal(right)
	if ok && err == nil {
		return Boolean(ok), nil
	}
	ok, err = cmp.Less(right)
	return Boolean(!ok), err
}
