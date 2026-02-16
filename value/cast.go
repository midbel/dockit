package value

import (
	"fmt"
)

type toFloat interface {
	ToFloat() (ScalarValue, error)
}

type toText interface {
	ToText() (ScalarValue, error)
}

func CastToArray(val Value) (Array, error) {
	arr, ok := val.(Array)
	if !ok {
		return arr, ErrCast
	}
	return arr, nil
}

func True(val Value) bool {
	b, ok := val.(Boolean)
	if ok {
		return bool(b)
	}
	return false
}

func CastToFloat(val Value) (Float, error) {
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

func CastToText(val Value) (Text, error) {
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
