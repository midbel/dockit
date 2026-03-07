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
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			res = min(res, float64(f))
		}
		return err
	})
	return value.Float(res), err
}

func Max(args []value.Value) (value.Value, error) {
	var res float64
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			res = max(res, float64(f))
		}
		return err
	})
	return value.Float(res), err
}

func Sum(args []value.Value) (value.Value, error) {
	var total float64
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			total += float64(f)
		}
		return err
	})
	return value.Float(total), err
}

func Avg(args []value.Value) (value.Value, error) {
	if len(args) == 0 {
		return value.Float(0), nil
	}
	var (
		total float64
		count int
	)
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			total += float64(f)
			count++
		}
		return err
	})
	if count == 0 {
		return value.ErrValue, nil
	}
	return value.Float(total / float64(count)), err
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
