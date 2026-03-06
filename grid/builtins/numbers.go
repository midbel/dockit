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
		if value.IsArray(args[i]) {
			sum, err := sumArray(args[i].(value.Array))
			if err != nil {
				return nil, err
			}
			total += sum
			continue
		}
		f, err := value.CastToFloat(args[i])
		if err != nil {
			return nil, err
		}
		total += float64(f)
	}
	return value.Float(total), nil
}

func Avg(args []value.Value) (value.Value, error) {
	if len(args) == 0 {
		return value.Float(0), nil
	}
	var (
		total float64
		count = len(args)
	)
	for i := range args {
		if value.IsArray(args[i]) {
			arr := args[i].(value.Array)
			count += int(arr.Count()) - 1
			sum, err := sumArray(arr)
			if err != nil {
				return nil, err
			}
			total += sum
			continue
		}
		if !value.IsNumber(args[i]) {
			return value.ErrValue, nil
		}
		total += float64(args[i].(value.Float))
	}
	return value.Float(total / float64(count)), nil
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

func sumArray(arr value.Array) (float64, error) {
	var total float64
	for v := range arr.Values() {
		v, err := value.CastToFloat(v)
		if err != nil {
			return total, err
		}
		total += float64(v)
	}
	return total, nil
}
