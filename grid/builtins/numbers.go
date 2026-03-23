package builtins

import (
	"math"

	"github.com/midbel/dockit/value"
)

func Sign(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
	if v < 0 {
		return value.Float(-1)
	}
	return value.Float(1)
}

func IsOdd(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
	return value.Boolean(math.Mod(float64(v), 2) == 1)
}

func IsEven(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
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
		if value.IsError(v) {
			return
		}
		ix++
		if f := asFloat(v); ix == 1 {
			res = f
		} else {
			res = min(res, f)
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
		if value.IsError(v) {
			return
		}
		res = max(res, asFloat(v))
	})
	if err := value.HasErrors(err); err != nil {
		return err
	}
	return value.Float(res)
}

func Sum(args []value.Value) value.Value {
	var total float64
	err := value.Each(args, func(v value.Value) {
		if value.IsError(v) {
			return
		}
		total += asFloat(v)
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
		if value.IsError(v) {
			return
		}
		total += asFloat(v)
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
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Round(f)
	return value.Float(ret)
}

func Floor(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Floor(f)
	return value.Float(ret)
}

func Ceil(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Ceil(f)
	return value.Float(ret)
}

func Sqrt(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	if f < 0 {
		return value.ErrValue
	}
	ret := math.Sqrt(f)
	return value.Float(ret)
}

func Abs(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	ret := math.Abs(f)
	return value.Float(ret)
}

func Mod(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		d = asFloat(args[1])
	)
	if d == 0 {
		return value.ErrDiv0
	}
	ret := math.Mod(f, d)
	return value.Float(ret)
}

func Pow(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		e = asFloat(args[1])
	)
	ret := math.Pow(f, e)
	return value.Float(ret)
}

func Int(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	f := asFloat(args[0])
	return value.Float(int(f))
}

func Rand(args []value.Value) value.Value {
	return nil
}

func Sin(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Sin(f)
	)
	return value.Float(r)
}

func Cos(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Cos(f)
	)
	return value.Float(r)
}

func Tan(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Tan(f)
	)
	return value.Float(r)
}

func Asin(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Asin(f)
	)
	return value.Float(r)
}

func Acos(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = math.Acos(f)
	)
	return value.Float(r)
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
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = f * (180 / math.Pi)
	)
	return value.Float(r)
}

func Rad(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	var (
		f = asFloat(args[0])
		r = f * (math.Pi / 180)
	)
	return value.Float(r)
}

func Pi(args []value.Value) value.Value {
	return value.Float(math.Pi)
}
