package types

import (
	"errors"
	"fmt"

	"github.com/midbel/dockit/value"
)

var ErrCast = errors.New("value can not be cast to target type")

type toFloat interface {
	ToFloat() (value.ScalarValue, error)
}

type toText interface {
	ToText() (value.ScalarValue, error)
}

func CastToFloat(val value.Value) (Float, error) {
	switch v := val.(type) {
	case Float:
		return v, nil
	case toFloat:
		x, err := v.ToFloat()
		if err != nil {
			return 0, ErrCast
		}
		f, ok := x.(Float)
		if !ok {
			return f, fmt.Errorf("cast error to float")
		}
		return f, nil
	default:
		return 0, ErrCast
	}
}

func CastToText(val value.Value) (Text, error) {
	switch v := val.(type) {
	case Text:
		return v, nil
	case toText:
		x, err := v.ToText()
		if err != nil {
			return "", ErrCast
		}
		f, ok := x.(Text)
		if !ok {
			return f, fmt.Errorf("cast error to text")
		}
		return f, nil
	default:
		return "", ErrCast
	}
}
