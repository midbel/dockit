package format

import (
	"testing"
	"time"

	"github.com/midbel/dockit/value"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		Pattern string
		Want    string
	}{
		{
			Pattern: "YYYY",
			Want:    "2026",
		},
		{
			Pattern: "YY",
			Want:    "26",
		},
		{
			Pattern: "MM",
			Want:    "2",
		},
		{
			Pattern: "0MM",
			Want:    "02",
		},
		{
			Pattern: "MMM",
			Want:    "Feb",
		},
		{
			Pattern: "MMMM",
			Want:    "February",
		},
		{
			Pattern: "DD",
			Want:    "20",
		},
		{
			Pattern: "0DD",
			Want:    "20",
		},
		{
			Pattern: "DDD",
			Want:    "Fri",
		},
		{
			Pattern: "DDDD",
			Want:    "Friday",
		},
		{
			Pattern: "JJJ",
			Want:    "51",
		},
		{
			Pattern: "0JJJ",
			Want:    "051",
		},
		{
			Pattern: "hh",
			Want:    "14",
		},
		{
			Pattern: "0mm",
			Want:    "05",
		},
		{
			Pattern: "0ss",
			Want:    "09",
		},
		{
			Pattern: "YYYY-MM-DD",
			Want:    "2026-2-20",
		},
		{
			Pattern: "YYYY-0MM-0DD",
			Want:    "2026-02-20",
		},
		{
			Pattern: "DD/MM/YYYY hh:mm:ss",
			Want:    "20/2/2026 14:5:9",
		},
		{
			Pattern: "0DD  MMM  YYYY",
			Want:    "20  Feb  2026",
		},
		{
			Pattern: "Report YYYY-MM-DD at 0hh:0mm",
			Want:    "Report 2026-2-20 at 14:05",
		},
	}

	today := time.Date(2026, 2, 20, 14, 5, 9, 0, time.UTC)
	for _, c := range tests {
		p, err := ParseDateFormatter(c.Pattern)
		if err != nil {
			t.Errorf("%s: error parsing pattern: %s", c.Pattern, err)
			continue
		}
		got, err := p.Format(value.Date(today))
		if err != nil {
			t.Errorf("%s: fail to format number (%v): %s", c.Pattern, today, err)
			continue
		}
		if got != c.Want {
			t.Errorf("%s (%v): results mismatched! want %s - got %s", c.Pattern, today, c.Want, got)
		}
	}
}
