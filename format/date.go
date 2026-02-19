package format

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/midbel/dockit/value"
)

func init() {
	slices.SortFunc(dateFieldsWriter, func(a, b dateFieldPattern) int {
		return strings.Compare(a.Pattern, b.Pattern)
	})
}

type dateFieldPattern struct {
	Pattern string
	Func    dateWriter
}

type dateWriter func(*strings.Builder, time.Time)

var dateFieldsWriter = []dateFieldPattern{
	{
		Pattern: "YYYY",
		Func:    writeYearLong,
	},
	{
		Pattern: "YY",
		Func:    writeYearShort,
	},
	{
		Pattern: "MM",
		Func:    writeMonth,
	},
	{
		Pattern: "0MM",
		Func:    writeMonthPadded,
	},
	{
		Pattern: "MMM",
		Func:    writeMonthNameShort,
	},
	{
		Pattern: "MMMM",
		Func:    writeMonthNameLong,
	},
	{
		Pattern: "DD",
		Func:    writeDay,
	},
	{
		Pattern: "0DD",
		Func:    writeDayPadded,
	},
	{
		Pattern: "DDD",
		Func:    writeDayNameShort,
	},
	{
		Pattern: "DDDD",
		Func:    writeDayNameLong,
	},
	{
		Pattern: "JJJ",
		Func:    writeYearDay,
	},
	{
		Pattern: "0JJJ",
		Func:    writeYearDayPadded,
	},
	{
		Pattern: "hh",
		Func:    writeHour,
	},
	{
		Pattern: "0hh",
		Func:    writeHourPadded,
	},
	{
		Pattern: "mm",
		Func:    writeMinute,
	},
	{
		Pattern: "0mm",
		Func:    writeMinutePadded,
	},
	{
		Pattern: "ss",
		Func:    writeSecond,
	},
	{
		Pattern: "0ss",
		Func:    writeSecondPadded,
	},
}

type dateFormatter struct {
	writers []dateWriter
}

func ParseDateFormatter(pattern string) (Formatter, error) {
	var df dateFormatter
	for i := 0; i < len(pattern); {
		var matched bool
		for _, k := range dateFieldsWriter {
			matched = strings.HasPrefix(pattern[i:], k.Pattern)
			if matched {
				df.writers = append(df.writers, k.Func)
				i += len(k.Pattern)
				matched = true
				break
			}
		}
		if !matched {
			i++
			df.writers = append(df.writers, writeLiteralDate(pattern[i]))
		}
	}
	return df, nil
}

func (f dateFormatter) Format(v value.Value) (string, error) {
	tv, ok := v.(value.Date)
	if !ok {
		return "", fmt.Errorf("value is not a date")
	}
	if len(f.writers) == 0 {
		return v.String(), nil
	}
	var str strings.Builder
	for i := range f.writers {
		f.writers[i](&str, time.Time(tv))
	}
	return str.String(), nil
}

func writeLiteralDate(char byte) dateWriter {
	return func(w *strings.Builder, _ time.Time) {
		w.WriteByte(char)
	}
}

func writeYearLong(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.Year()))
}

func writeYearShort(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.Year() % 100))
}

func writeMonth(w *strings.Builder, t time.Time) {
	m := int(t.Month())
	w.WriteString(strconv.Itoa(m))
}

func writeMonthPadded(w *strings.Builder, t time.Time) {
	m := int(t.Month())
	if m < 10 {
		w.WriteByte('0')
	}
	w.WriteString(strconv.Itoa(m))
}

func writeMonthNameShort(w *strings.Builder, t time.Time) {

}

func writeMonthNameLong(w *strings.Builder, t time.Time) {

}

func writeDay(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.Day()))
}

func writeDayPadded(w *strings.Builder, t time.Time) {
	d := t.Day()
	if d < 10 {
		w.WriteByte('0')
	}
	w.WriteString(strconv.Itoa(d))
}

func writeDayNameShort(w *strings.Builder, t time.Time) {

}

func writeDayNameLong(w *strings.Builder, t time.Time) {

}

func writeYearDay(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.YearDay()))
}

func writeYearDayPadded(w *strings.Builder, t time.Time) {
	y := t.Day()
	if y < 10 {
		w.WriteByte('0')
	}
	if y < 100 {
		w.WriteByte('0')
	}
	w.WriteString(strconv.Itoa(y))
}

func writeHour(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.Hour()))
}

func writeHourPadded(w *strings.Builder, t time.Time) {
	h := t.Hour()
	if h < 10 {
		w.WriteByte('0')
	}
	w.WriteString(strconv.Itoa(h))
}

func writeMinute(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.Minute()))
}

func writeMinutePadded(w *strings.Builder, t time.Time) {
	m := t.Minute()
	if m < 10 {
		w.WriteByte('0')
	}
	w.WriteString(strconv.Itoa(m))
}

func writeSecond(w *strings.Builder, t time.Time) {
	w.WriteString(strconv.Itoa(t.Second()))
}

func writeSecondPadded(w *strings.Builder, t time.Time) {
	s := t.Second()
	if s < 10 {
		w.WriteByte('0')
	}
	w.WriteString(strconv.Itoa(s))
}
