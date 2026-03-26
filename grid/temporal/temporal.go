package temporal

import (
	"time"
)

func YearsBetween(dtend, dtstart time.Time) int {
	var year int

	dtstart = dtstart.UTC()
	dtend = dtend.UTC()
	for dtstart.Before(dtend) {
		dtstart = dtstart.AddDate(1, 0, 0)
		year++
	}
	return year - 1
}

func MonthsBetween(dtend, dtstart time.Time) int {
	var month int

	dtstart = dtstart.UTC()
	dtend = dtend.UTC()
	for dtstart.Before(dtend) {
		dtstart = dtstart.AddDate(0, 1, 0)
		month++
	}
	return month - 1
}

func DaysBetween(dtend, dtstart time.Time) int {
	dtstart = dtstart.UTC()
	dtend = dtend.UTC()

	days := dtend.Sub(dtstart)
	return int(days.Abs().Hours()) / 24
}

func CountDays(dtend, dtstart time.Time) int {
	dtstart = dtstart.UTC()
	dtend = dtend.UTC()

	dt := time.Date(dtend.Year(), dtstart.Month(), dtstart.Day(), dtstart.Hour(), dtstart.Minute(), dtstart.Second(), 0, time.UTC)
	return DaysBetween(dtend, dt)
}

func CountDays2(dtend, dtstart time.Time) int {
	return 0
}

func CountMonths(dtend, dtstart time.Time) int {
	dtstart = dtstart.UTC()
	dtend = dtend.UTC()

	dt := time.Date(dtend.Year(), dtstart.Month(), dtstart.Day(), dtstart.Hour(), dtstart.Minute(), dtstart.Second(), 0, time.UTC)
	return MonthsBetween(dtend, dt)
}
