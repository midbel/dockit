package format

import (
	"testing"
	"time"

	"github.com/midbel/dockit/value"
)

func TestFormatter(t *testing.T) {
	tests := []struct {
		Input value.Value
		Want  string
	}{
		{
			Input: value.Float(42),
			Want:  "42",
		},
		{
			Input: value.Float(123),
			Want:  "123",
		},

		{
			Input: value.Float(3.14),
			Want:  "3.14",
		},
		{
			Input: value.Date(time.Date(2026, 2, 20, 14, 5, 9, 0, time.UTC)),
			Want:  "2026-02-20",
		},
		{
			Input: value.Text("foobar"),
			Want:  "foobar",
		},
		{
			Input: value.Boolean(true),
			Want:  "true",
		},
	}
	vf := FormatValue()
	vf.Number("###.##")
	vf.Date("YYYY-0MM-0DD")
	vf.Set(value.TypeText, FormatString())
	vf.Set(value.TypeBool, FormatBool())
	for _, c := range tests {
		got, err := vf.Format(c.Input)
		if err != nil {
			t.Errorf("fail to format value (%v): %s", c.Input, err)
			continue
		}
		if got != c.Want {
			t.Errorf("%v: results mismatched! want %s - got %s", c.Input, c.Want, got)
		}
	}
}
