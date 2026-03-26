package builtins

import (
	"time"

	"github.com/midbel/dockit/grid/temporal"
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
	if err := value.HasErrors(args[:3]...); err != nil {
		return err
	}
	var (
		year  = asFloat(args[0])
		month = asFloat(args[1])
		day   = asFloat(args[2])
	)
	n := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return value.Date(n)
}

func Year(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Year())
}

func Month(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Month())
}

func Day(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Day())
}

func YearDay(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.YearDay())
}

func Hour(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Hour())
}

func Minute(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Minute())
}

func Second(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Second())
}

func Weekday(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Weekday())
}

func DateDiff(args []value.Value) value.Value {
	var (
		dtstart = asTime(args[0])
		dtend   = asTime(args[1])
		unit    = asString(args[2])
		delta   float64
	)
	if dtstart.After(dtend) {
		return value.ErrNum
	}
	switch unit {
	case "Y":
		diff := temporal.YearsBetween(dtend, dtstart)
		delta = float64(diff)
	case "M":
		diff := temporal.MonthsBetween(dtend, dtstart)
		delta = float64(diff)
	case "D":
		diff := temporal.DaysBetween(dtend, dtstart)
		delta = float64(diff)
	case "YD":
		diff := temporal.CountDays(dtend, dtstart)
		delta = float64(diff)
	case "YM":
		diff := temporal.CountMonths(dtend, dtstart)
		delta = float64(diff)
	case "MD":
	default:
	}
	return value.Float(delta)
}
