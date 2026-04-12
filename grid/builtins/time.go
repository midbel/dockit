package builtins

import (
	"time"

	"github.com/midbel/dockit/grid/temporal"
	"github.com/midbel/dockit/value"
)

var nowBuiltin = Builtin{
	Name:     "now",
	Desc:     "Returns the current date and time",
	Category: "time",
	Func:     Now,
}

func Now(args []value.Value) value.Value {
	n := time.Now()
	return value.Date(n)
}

var todayBuiltin = Builtin{
	Name:     "today",
	Desc:     "Returns the current date and time",
	Category: "time",
	Func:     Today,
}

func Today(args []value.Value) value.Value {
	n := time.Now().Truncate(time.Hour * 24)
	return value.Date(n)
}

var dateBuiltin = Builtin{
	Name:     "date",
	Desc:     "Creates a date from year, month, and day",
	Category: "time",
	Params: []Param{
		Scalar("year", "", value.TypeNumber),
		Scalar("month", "", value.TypeNumber),
		Scalar("day", "", value.TypeNumber),
	},
	Func: Date,
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

var yearBuiltin = Builtin{
	Name:     "year",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Year,
}

func Year(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Year())
}

var monthBuiltin = Builtin{
	Name:     "month",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Month,
}

func Month(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Month())
}

var dayBuiltin = Builtin{
	Name:     "day",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Day,
}

func Day(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Day())
}

var yeardayBuiltin = Builtin{
	Name:     "yearday",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: YearDay,
}

func YearDay(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.YearDay())
}

var hourBuiltin = Builtin{
	Name:     "hour",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Hour,
}

func Hour(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Hour())
}

var minuteBuiltin = Builtin{
	Name:     "minute",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Minute,
}

func Minute(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Minute())
}

var secondBuiltin = Builtin{
	Name:     "second",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Second,
}

func Second(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Second())
}

var weekdayBuiltin = Builtin{
	Name:     "weekday",
	Desc:     "",
	Category: "time",
	Params: []Param{
		Scalar("date", "", value.TypeDate),
	},
	Func: Weekday,
}

func Weekday(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	t := asTime(args[0])
	return value.Float(t.Weekday())
}

func Edate(args []value.Value) value.Value {
	return nil
}

func EoMonth(args []value.Value) value.Value {
	return nil
}

var datediffBuiltin = Builtin{
	Name:     "datedif",
	Desc:     "Returns the difference between two dates",
	Category: "time",
	Params: []Param{
		Scalar("fromDate", "", value.TypeDate),
		Scalar("toDate", "", value.TypeDate),
		Scalar("diffUnit", "", value.TypeText),
	},
	Func: DateDiff,
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

var timeBuiltins = []Builtin{
	nowBuiltin,
	todayBuiltin,
	dateBuiltin,
	yearBuiltin,
	monthBuiltin,
	dayBuiltin,
	yeardayBuiltin,
	hourBuiltin,
	minuteBuiltin,
	secondBuiltin,
	weekdayBuiltin,
	datediffBuiltin,
}
