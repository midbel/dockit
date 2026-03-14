package builtins

import (
	"time"

	"github.com/midbel/dockit/value"
)

func Now(args []value.Value) value.Value {
	n := time.Now()
	return value.Date(n)
}

func Today(args []value.Value) value.Value {
	n := time.Now().Truncate(time.Hour * 24)
	return value.Date(n)
}

func Date(args []value.Value) value.Value {
	year, err := value.CastToFloat(args[0])
	if err != nil {
		return value.ErrValue
	}
	month, err := value.CastToFloat(args[1])
	if err != nil {
		return value.ErrValue
	}
	day, err := value.CastToFloat(args[2])
	if err != nil {
		return value.ErrValue
	}

	n := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return value.Date(n)
}

func Year(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Year())
}

func Month(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Month())
}

func Day(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Day())
}

func YearDay(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.YearDay())
}

func Hour(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Hour())
}

func Minute(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Minute())
}

func Second(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Second())
}

func Weekday(args []value.Value) value.Value {
	d, err := value.CastToDate(args[0])
	if err != nil {
		return value.ErrValue
	}
	t := time.Time(d)
	return value.Float(t.Weekday())
}
