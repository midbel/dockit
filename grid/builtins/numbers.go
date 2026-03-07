package builtins

import (
	"math"

	"github.com/midbel/dockit/value"
)

func IsNumber(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	ok := value.IsNumber(args[0])
	return value.Boolean(ok), nil
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
	return value.Float(0), nil
}

func Round(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := math.Round(float64(f))
	return value.Float(ret), nil
}

func Floor(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := math.Floor(float64(f))
	return value.Float(ret), nil
}

func Ceil(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := math.Ceil(float64(f))
	return value.Float(ret), nil
}

func Sqrt(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := math.Sqrt(float64(f))
	return value.Float(ret), nil
}

func Abs(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	ret := math.Abs(float64(f))
	return value.Float(ret), nil
}

func Mod(args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	d, err := value.CastToFloat(args[1])
	if err != nil {
		return value.ErrValue, err
	}
	if d == 0 {
		return value.ErrDiv0, nil
	}
	ret := math.Mod(float64(f), float64(d))
	return value.Float(ret), nil
}

func Pow(args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return value.ErrValue, ErrArity
	}
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	e, err := value.CastToFloat(args[1])
	if err != nil {
		return value.ErrValue, err
	}
	ret := math.Pow(float64(f), float64(e))
	return value.Float(ret), nil
}

func Int(args []value.Value) (value.Value, error) {
	return value.Float(0), nil
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
