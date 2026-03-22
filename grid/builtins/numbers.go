package builtins

import (
	"math"

	"github.com/midbel/dockit/value"
)

func Sign(args []value.Value) value.Value {
	v, _ := value.CastToFloat(args[0])
	if v < 0 {
		return value.Float(-1)
	}
	return value.Float(1)
}

func IsOdd(args []value.Value) value.Value {
	v, _ := value.CastToFloat(args[0])
	return value.Boolean(math.Mod(float64(v), 2) == 1)
}

func IsEven(args []value.Value) value.Value {
	v, _ := value.CastToFloat(args[0])
	return value.Boolean(math.Mod(float64(v), 2) == 0)
}

func IsNumber(args []value.Value) value.Value {
	ok := value.IsNumber(args[0])
	return value.Boolean(ok)
}

func Min(args []value.Value) value.Value {
	var (
		res float64
		ix  int
	)
	err := value.Each(args, func(v value.Value) {
		f, err := value.CastToFloat(v)
		if err != nil {
			return
		}
		ix++
		if ix == 1 {
			res = float64(f)
		} else {
			res = min(res, float64(f))
		}
	})
	if err := value.HasErrors(err); err != nil {
		return err
	}
	return value.Float(res)
}

func Max(args []value.Value) value.Value {
	var res float64
	err := value.Each(args, func(v value.Value) {
		f, err := value.CastToFloat(v)
		if err != nil {
			return
		}
		res = max(res, float64(f))
	})
	if err := value.HasErrors(err); err != nil {
		return err
	}
	return value.Float(res)
}

func Sum(args []value.Value) value.Value {
	var total float64
	err := value.Each(args, func(v value.Value) {
		f, err := value.CastToFloat(v)
		if err != nil {
			return
		}
		total += float64(f)
	})
	if err := value.HasErrors(err); err != nil {
		return err
	}
	return value.Float(total)
}

func Avg(args []value.Value) value.Value {
	var (
		total float64
		count int
	)
	err := value.Each(args, func(v value.Value) {
		f, err := value.CastToFloat(v)
		if err != nil {
			return
		}
		total += float64(f)
		count++
	})
	if err := value.HasErrors(err); err != nil {
		return err
	}
	if count == 0 {
		return value.ErrDiv0
	}
	return value.Float(total / float64(count))
}

func Stdev(args []value.Value) value.Value {
	return nil
}

func Variance(args []value.Value) value.Value {
	return nil
}

func Mode(args []value.Value) value.Value {
	return nil
}

func Median(args []value.Value) value.Value {
	return nil
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
	if float64(f) < 0 {
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
	f, _ := value.CastToFloat(args[0])
	return value.Float(int(f))
}

func Rand(args []value.Value) value.Value {
	return nil
}

func Sin(args []value.Value) value.Value {
	var (
		val, _ = value.CastToFloat(args[0])
		ret    = math.Sin(float64(val))
	)
	return value.Float(ret)
}

func Cos(args []value.Value) value.Value {
	var (
		val, _ = value.CastToFloat(args[0])
		ret    = math.Cos(float64(val))
	)
	return value.Float(ret)
}

func Tan(args []value.Value) value.Value {
	var (
		val, _ = value.CastToFloat(args[0])
		ret    = math.Tan(float64(val))
	)
	return value.Float(ret)
}

func Asin(args []value.Value) value.Value {
	var (
		val, _ = value.CastToFloat(args[0])
		ret    = math.Asin(float64(val))
	)
	return value.Float(ret)
}

func Acos(args []value.Value) value.Value {
	var (
		val, _ = value.CastToFloat(args[0])
		ret    = math.Acos(float64(val))
	)
	return value.Float(ret)
}

func Atan2(args []value.Value) value.Value {
	var (
		vx, _ = value.CastToFloat(args[0])
		vy, _ = value.CastToFloat(args[1])
		ret   = math.Atan2(float64(vx), float64(vy))
	)
	return value.Float(ret)
}

func Deg(args []value.Value) value.Value {
	var (
		g, _ = value.CastToFloat(args[0])
		res  = float64(g) * (180 / math.Pi)
	)
	return value.Float(res)
}

func Rad(args []value.Value) value.Value {
	var (
		g, _ = value.CastToFloat(args[0])
		res  = float64(g) * (math.Pi / 180)
	)
	return value.Float(res)
}

func Pi(args []value.Value) value.Value {
	return value.Float(math.Pi)
}
