package builtins

import (
	"math"

	"github.com/midbel/dockit/value"
)

func IsNumber(args []value.Value) value.Value {
	ok := value.IsNumber(args[0])
	return value.Boolean(ok)
}

func Min(args []value.Value) value.Value {
	var (
		res float64
		ix  int
	)
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			ix++
			if ix == 1 {
				res = float64(f)
			} else {
				res = min(res, float64(f))
			}
		}
		return err
	})
	if err != nil {
		return value.ErrValue
	}
	return value.Float(res)
}

func Max(args []value.Value) value.Value {
	var res float64
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			res = max(res, float64(f))
		}
		return err
	})
	if err != nil {
		return value.ErrValue
	}
	return value.Float(res)
}

func Sum(args []value.Value) value.Value {
	var total float64
	err := Each(args, func(v value.Value) error {
		f, err := value.CastToFloat(v)
		if err == nil {
			total += float64(f)
		}
		return err
	})
	if err != nil {
		return value.ErrValue
	}
	return value.Float(total)
}

func Avg(args []value.Value) value.Value {
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
	if err != nil {
		return value.ErrValue
	}
	if count == 0 {
		return value.ErrValue
	}
	return value.Float(total / float64(count))
}

func Count(args []value.Value) value.Value {
	return value.Float(0)
}

func Round(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := math.Round(float64(f))
	return value.Float(ret)
}

func Floor(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := math.Floor(float64(f))
	return value.Float(ret)
}

func Ceil(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := math.Ceil(float64(f))
	return value.Float(ret)
}

func Sqrt(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := math.Sqrt(float64(f))
	return value.Float(ret)
}

func Abs(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	ret := math.Abs(float64(f))
	return value.Float(ret)
}

func Mod(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	d, err := value.CastToFloat(args[1])
	if err != nil {
		return value.ErrValue
	}
	if d == 0 {
		return value.ErrDiv0
	}
	ret := math.Mod(float64(f), float64(d))
	return value.Float(ret)
}

func Pow(args []value.Value) value.Value {
	f, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	e, err := value.CastToFloat(args[1])
	if err != nil {
		return value.ErrValue
	}
	ret := math.Pow(float64(f), float64(e))
	return value.Float(ret)
}

func Int(args []value.Value) value.Value {
	return value.Float(0)
}

func Rand(args []value.Value) value.Value {
	return nil
}
