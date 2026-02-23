package builtins

import (
	"github.com/midbel/dockit/value"
)

func IsNumber(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Min(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !value.IsNumber(args[i]) {
			return value.ErrValue, nil
		}
		v := args[i].(value.Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = min(res, float64(v))
	}
	return value.Float(res), nil
}

func Max(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !value.IsNumber(args[i]) {
			return value.ErrValue, nil
		}
		v := args[i].(value.Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = max(res, float64(v))
	}
	return value.Float(res), nil
}

func Sum(args []value.Value) (value.Value, error) {
	var total float64
	for i := range args {
		if !value.IsNumber(args[i]) {
			return value.ErrValue, nil
		}
		total += float64(args[i].(value.Float))
	}
	return value.Float(total), nil
}

func Avg(args []value.Value) (value.Value, error) {
	if len(args) == 0 {
		return value.Float(0), nil
	}
	var total float64
	for i := range args {
		if !value.IsNumber(args[i]) {
			return value.ErrValue, nil
		}
		total += float64(args[i].(value.Float))
	}
	return value.Float(total / float64(len(args))), nil
}

func Count(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Round(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Floor(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Ceil(args []value.Value) (value.Value, error) {
	return nil, nil
}

func Sqrt(args []value.Value) (value.Value, error) {
	return nil, nil
}
