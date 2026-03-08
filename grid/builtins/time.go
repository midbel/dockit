package builtins

import (
	"time"

	"github.com/midbel/dockit/value"
)

func Now(args []value.Value) (value.Value, error) {
	n := time.Now()
	return value.Date(n), nil
}

func Today(args []value.Value) (value.Value, error) {
	n := time.Now().Truncate(time.Hour * 24)
	return value.Date(n), nil
}

func Date(args []value.Value) (value.Value, error) {
	year, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	month, err := value.CastToFloat(args[1])
	if err != nil {
		return value.ErrValue, err
	}
	day, err := value.CastToFloat(args[2])
	if err != nil {
		return value.ErrValue, err
	}

	n := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return value.Date(n), nil
}

func Year(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Year()), nil
}

func Month(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Month()), nil
}

func Day(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Day()), nil
}

func YearDay(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.YearDay()), nil
}

func Hour(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Hour()), nil
}

func Minute(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Minute()), nil
}

func Second(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Second()), nil
}

func Weekday(args []value.Value) (value.Value, error) {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue, err
	}
	t := time.Time(d)
	return value.Float(t.Weekday()), nil
}
