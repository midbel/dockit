package format

import (
	"testing"

	"github.com/midbel/dockit/formula/parse"
)

func TestFormat(t *testing.T) {
	t.Run("ods", testOds)
	t.Run("oxml", testOxml)
	t.Run("cross-format", testCrossFormat)
}

func testCrossFormat(t *testing.T) {
	tests := []struct {
		Expr      string
		Want      string
		Parse     func(string) (parse.Expr, error)
		Formatter DialectFormat
	}{
		{
			Expr:      "=A1",
			Want:      "of:=[.A1]",
			Parse:     parse.ParseOxmlFormula,
			Formatter: Ods,
		},
		{
			Expr:      "of:=[.A1]",
			Want:      "=A1",
			Parse:     parse.ParseOdsFormula,
			Formatter: Oxml,
		},
	}
	for _, c := range tests {
		e, err := c.Parse(c.Expr)
		if err != nil {
			t.Errorf("%s: expression error: %s", c.Expr, err)
			continue
		}
		got, err := Format(e, c.Formatter)
		if err != nil {
			t.Errorf("%s: error formatting expression: %s", c.Expr, err)
			continue
		}
		if c.Want != got {
			t.Errorf("%s: results mismatched! want %s, got %s", c.Expr, c.Want, got)
		}
	}
}

func testOds(t *testing.T) {
	tests := []string{
		"of:=[.A1]",
		"of:=[.A1] & \"test\"",
		"of:=sum(1;2;3;[.A1:.A100])",
	}
	for _, c := range tests {
		e, err := parse.ParseOdsFormula(c)
		if err != nil {
			t.Errorf("%s: expression error: %s", c, err)
			continue
		}
		got, err := FormatOds(e)
		if err != nil {
			t.Errorf("%s: error formatting expression: %s", c, err)
			continue
		}
		if c != got {
			t.Errorf("results mismatched! want %s, got %s", c, got)
		}
	}
}

func testOxml(t *testing.T) {
	tests := []string{
		"=A1",
	}
	for _, c := range tests {
		e, err := parse.ParseOxmlFormula(c)
		if err != nil {
			t.Errorf("expression error: %s", err)
			continue
		}
		got, err := FormatOxml(e)
		if err != nil {
			t.Errorf("error formatting expression: %s", err)
			continue
		}
		if c != got {
			t.Errorf("results mismatched! want %s, got %s", c, got)
		}
	}
}
