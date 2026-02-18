package eval

import (
	"testing"

	"github.com/midbel/dockit/value"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		Pattern string
		Input   value.Value
		Want    string
	}{
		{
			Pattern: "###",
			Input:   value.Float(42),
			Want:    "42",
		},
		{
			Pattern: "0000",
			Input:   value.Float(42),
			Want:    "0042",
		},
		{
			Pattern: "0.00",
			Input:   value.Float(1.234),
			Want:    "1.23",
		},
		{
			Pattern: "0.00",
			Input:   value.Float(1.999),
			Want:    "2.00",
		},
		{
			Pattern: "0.0#",
			Input:   value.Float(1.20),
			Want:    "1.2", // strip trailing zero beyond minDec
		},
		{
			Pattern: "0.0#",
			Input:   value.Float(1.25),
			Want:    "1.25", // keep all significant digits
		},
		{
			Pattern: "0.##",
			Input:   value.Float(1.20),
			Want:    "1.2", // optional decimals trimmed beyond first significant
		},
		{
			Pattern: "0.##",
			Input:   value.Float(1.25),
			Want:    "1.25",
		},
		{
			Pattern: "0",
			Input:   value.Float(1.6),
			Want:    "2",
		},
		{
			Pattern: "#,###",
			Input:   value.Float(12345),
			Want:    "12,345",
		},
		{
			Pattern: "#,###.00",
			Input:   value.Float(12345.2),
			Want:    "12,345.20",
		},
		{
			Pattern: "0.00",
			Input:   value.Float(-12.3),
			Want:    "-12.30",
		},
		{
			Pattern: "+0.00",
			Input:   value.Float(12.3),
			Want:    "+12.30",
		},
		{
			Pattern: "+0.00",
			Input:   value.Float(-12.3),
			Want:    "-12.30",
		},
		{
			Pattern: "000,000",
			Input:   value.Float(1234),
			Want:    "001,234",
		},
		{
			Pattern: "0.00",
			Input:   value.Float(0),
			Want:    "0.00",
		},
	}
	for _, c := range tests {
		p, err := ParseNumberFormatter(c.Pattern)
		if err != nil {
			t.Errorf("%s: error parsing pattern: %s", c.Pattern, err)
			continue
		}
		got, err := p.Format(c.Input)
		if err != nil {
			t.Errorf("%s: fail to format number (%v): %s", c.Pattern, c.Input, err)
			continue
		}
		if got != c.Want {
			t.Errorf("%s (%v): results mismatched! want %s - got %s", c.Pattern, c.Input, c.Want, got)
		}
	}
}
