package builtins

import (
	"math"

	"github.com/midbel/dockit/grid/calc"
	"github.com/midbel/dockit/grid/criteria"
	"github.com/midbel/dockit/value"
)

func Sign(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	s := calc.Sign(asFloat(args[0]))
	return value.Float(s)
}

func IsOdd(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
	return value.Boolean(calc.Odd(v))
}

func IsEven(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	v := asFloat(args[0])
	return value.Boolean(calc.Even(v))
}

func IsNumber(args []value.Value) value.Value {
	ok := value.IsNumber(args[0])
	return value.Boolean(ok)
}

func Min(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Min(arr)
	)
	return value.Float(ret)
}

func Max(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Max(arr)
	)
	return value.Float(ret)
}

func Sum(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Sum(arr)
	)
	return value.Float(ret)
}

func SumIf(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f, err := criteria.New(asString(args[1]))
	if err != nil {
		return value.ErrValue
	}
	total := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		if ok := f.Keep(v); ok {
			acc += asFloat(v)
		}
		return acc
	})
	return value.Float(total)
}

func Avg(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Avg(arr)
	)
	return value.Float(ret)
}

func AvgIf(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f, err := criteria.New(asString(args[1]))
	if err != nil {
		return value.ErrValue
	}
	vs := value.Collect[float64](args, func(v value.Value) (float64, bool) {
		return asFloat(v), f.Keep(v)
	})
	ret := calc.Avg(vs)
	return value.Float(ret)
}

func Stdev(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Stdev(arr)
	)
	return value.Float(ret)
}

func Variance(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Var(arr)
	)
	return value.Float(ret)
}

func Mode(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Mode(arr)
	)
	return value.Float(ret)
}

func Median(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		arr = asFloatArray(args)
		ret = calc.Median(arr)
	)
	return value.Float(ret)
}

func Count(args []value.Value) value.Value {
	count := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		return acc + 1
	})
	return value.Float(count)
}

func CountIf(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f, err := criteria.New(asString(args[1]))
	if err != nil {
		return value.ErrValue
	}
	count := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		if ok := f.Keep(v); ok {
			acc += 1
		}
		return acc
	})
	return value.Float(count)
}

func Counta(args []value.Value) value.Value {
	count := value.Reduce[float64](args, 0, func(acc float64, v value.Value) float64 {
		if value.IsBlank(v) || asString(v) == "" {
			return acc
		}
		return acc + 1
	})
	return value.Float(count)
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
	r := calc.Rand()
	return value.Float(r)
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
		vx  = asFloat(args[0])
		vy  = asFloat(args[1])
		ret = math.Atan2(vx, vy)
	)
	return value.Float(ret)
}

func Deg(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	r := calc.Deg(asFloat(args[0]))
	return value.Float(r)
}

func Rad(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	r := calc.Rad(asFloat(args[0]))
	return value.Float(r)
}

func Pi(args []value.Value) value.Value {
	pi := calc.Pi()
	return value.Float(pi)
}
