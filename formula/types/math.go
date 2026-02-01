package types

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/value"
)

var ErrOperation = errors.New("operation not supported")

func Add(left, right value.Value) (value.Value, error) {
	a, ok := left.(interface {
		Add(value.Value) (value.ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s + %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Add(right)
}

func Sub(left, right value.Value) (value.Value, error) {
	a, ok := left.(interface {
		Sub(value.Value) (value.ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s - %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Sub(right)
}

func Mul(left, right value.Value) (value.Value, error) {
	a, ok := left.(interface {
		Mul(value.Value) (value.ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s * %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Mul(right)
}

func Div(left, right value.Value) (value.Value, error) {
	a, ok := left.(interface {
		Div(value.Value) (value.ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s / %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Div(right)
}

func Pow(left, right value.Value) (value.Value, error) {
	a, ok := left.(interface {
		Pow(value.Value) (value.ScalarValue, error)
	})
	if !ok {
		return ErrValue, fmt.Errorf("%w: %s ^ %s", ErrOperation, left.Type(), right.Type())
	}
	return a.Pow(right)
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
