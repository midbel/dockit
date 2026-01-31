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
