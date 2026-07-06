package builtins

import (
	"testing"
	"time"

	"github.com/midbel/dockit/value"
)

var today = time.Date(2026, 6, 6, 15, 25, 0, 0, time.UTC)

func TestDates(t *testing.T) {
	t.Run("date", testDate)
	t.Run("year", testYear)
	t.Run("month", testMonth)
	t.Run("yearday", testYearDay)
	t.Run("day", testDay)
	t.Run("hour", testHour)
	t.Run("minute", testMinute)
	t.Run("second", testSecond)
	t.Run("weekday", testWeekDay)
	t.Run("datediff", testDateDiff)
}

func testDate(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(2026),
				value.Float(6),
				value.Float(6),
			},
			Want: value.Date(today.Truncate(time.Hour * 24)),
		},
	}
	testBuiltin(t, Date, tests)
}

func testYear(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(2026),
		},
	}
	testBuiltin(t, Year, tests)
}

func testMonth(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(6),
		},
	}
	testBuiltin(t, Month, tests)
}

func testDay(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(6),
		},
	}
	testBuiltin(t, Day, tests)
}

func testYearDay(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(157),
		},
	}
	testBuiltin(t, YearDay, tests)
}

func testHour(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(15),
		},
	}
	testBuiltin(t, Hour, tests)
}

func testMinute(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(25),
		},
	}
	testBuiltin(t, Minute, tests)
}

func testSecond(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(0),
		},
	}
	testBuiltin(t, Second, tests)
}

func testWeekDay(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Date(today)},
			Want: value.Float(6),
		},
	}
	testBuiltin(t, Weekday, tests)
}

func testDateDiff(t *testing.T) {
	t.SkipNow()
}
