package grid_test

import (
	"testing"

	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

func TestFormula(t *testing.T) {
	tests := []struct {
		Formula string
		Want    string
	}{
		{
			Formula: "=B1+B2",
			Want:    "7",
		},
		{
			Formula: "=sheet1!B1 + sheet1!B2",
			Want:    "7",
		},
		{
			Formula: "=-B1+B2",
			Want:    "3",
		},
		{
			Formula: "=-sheet2!B2",
			Want:    "-5",
		},
		{
			Formula: "=B1*B2",
			Want:    "10",
		},
		{
			Formula: "=B1 * sheet2!B1",
			Want:    "20",
		},
		{
			Formula: "=A1&A2",
			Want:    "foobar",
		},
		{
			Formula: "=A1 & ' ' & sheet2!A1",
			Want:    "foo quz",
		},
		{
			Formula: "=MIN(B1:B2)",
			Want:    "2",
		},
		{
			Formula: "=MAX(B1:B2)",
			Want:    "5",
		},
		{
			Formula: "=UPPER(sheet2!A1)",
			Want: "QUZ",
		},
	}
	ctx := getContext()
	for _, c := range tests {
		val, err := grid.EvalString(c.Formula, ctx)
		if err != nil {
			t.Errorf("%s: error executing formula: %s", c.Formula, err)
			continue
		}
		if got := val.String(); got != c.Want {
			t.Errorf("%s: result mismatched! want %s, got %s", c.Formula, c.Want, got)
		}
	}
}

func getContext() value.Context {
	sheet1 := flat.NewSheet("sheet1", value.Rows(
		[]value.ScalarValue{value.Text("foo"), value.Float(2)},
		[]value.ScalarValue{value.Text("bar"), value.Float(5)},
	))

	sheet2 := flat.NewSheet("sheet2", value.Rows(
		[]value.ScalarValue{value.Text("quz"), value.Float(10)},
		[]value.ScalarValue{value.Text("bee"), value.Float(5)},
	))

	file := flat.NewFileFromSheets(sheet1, sheet2)

	return grid.FileContext(file)
}
