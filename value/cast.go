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

type toBool interface {
	ToBool() (ScalarValue, error)
}

type toDate interface {
	ToDate() (ScalarValue, error)
}

func CastToArray(val Value) (Array, error) {
	arr, ok := val.(Array)
	if !ok {
		return arr, ErrCast
	}
	return arr, nil
}

func True(val Value) bool {
	tb, ok := val.(toBool)
	if !ok {
		return false
	}
	b, err := tb.ToBool()
	if err != nil {
		return false
	}
	if b, ok := b.(Boolean); ok {
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
			return 0, errorCast(val, TypeNumber)
		}
		f, ok := x.(Float)
		if !ok {
			return f, ErrCast
		}
		return f, nil
	default:
		return 0, errorCast(val, TypeNumber)
	}
}

func CastToText(val Value) (Text, error) {
	switch v := val.(type) {
	case Text:
		return v, nil
	case toText:
		x, err := v.ToText()
		if err != nil {
			return "", errorCast(val, TypeText)
		}
		f, ok := x.(Text)
		if !ok {
			return f, ErrCast
		}
		return f, nil
	default:
		return "", errorCast(val, TypeText)
	}
}

func CastToDate(val Value) (Date, error) {
	switch v := val.(type) {
	case Date:
		return v, nil
	case toDate:
		x, err := v.ToDate()
		if err != nil {
			var z Date
			return z, errorCast(val, TypeDate)
		}
		f, ok := x.(Date)
		if !ok {
			return f, ErrCast
		}
		return f, nil
	default:
		var z Date
		return z, errorCast(val, TypeDate)
	}
}

func errorCast(val Value, target string) error {
	return fmt.Errorf("%s(%s): %w (%s)", val.Type(), val.String(), ErrCast, target)
}
