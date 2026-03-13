package grid_test

import (
	"testing"

	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

type FormulaTestCase struct {
	Formula string
	Want    string
}

func TestEvalErrors(t *testing.T) {
	tests := []struct {
		Comment string
		Formula string
		Want    string
	}{
		{
			Comment: "division by zero",
			Formula: "=1/0",
			Want:    "#DIV/0!",
		},
		{
			Comment: "incompatible types",
			Formula: "=\"foo\"+1",
			Want:    "#VALUE!",
		},
		{
			Comment: "references error",
			Formula: "=Z9999",
			Want:    "#REF!",
		},
		{
			Comment: "unknown function",
			Formula: "=FOOBAR(1)",
			Want:    "#NAME?",
		},
	}
	ctx := getContext()
	for _, c := range tests {
		val, err := grid.EvalString(c.Formula, ctx)
		if err != nil {
			t.Errorf("%s: error executing formula: %s", c.Comment, err)
			continue
		}
		if val.Type() != value.TypeError {
			t.Errorf("%s: expected error type, got %s", c.Comment, val.Type())
			continue
		}
		if got := val.String(); got != c.Want {
			t.Errorf("%s: result mismatched! want %s, got %s", c.Comment, c.Want, got)
		}
	}
}

func TestBasicFormula(t *testing.T) {
	tests := []FormulaTestCase{
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
			Formula: "=2^2",
			Want:    "4",
		},
	}
	runTests(t, tests)
}

func TestBuiltins(t *testing.T) {
	t.Run("arithmetic", testMathBuiltins)
	t.Run("text", testStringBuiltins)
}

func testStringBuiltins(t *testing.T) {
	tests := []FormulaTestCase{
		{
			Formula: "=UPPER(sheet2!A1)",
			Want:    "QUZ",
		},
		{
			Formula: "=lower(UPPER(sheet2!A1))",
			Want:    "quz",
		},
		{
			Formula: "concat(A1, A2, sheet2!A1, sheet2!A2)",
			Want:    "foobarquzbee",
		},
		{
			Formula: "len('foobar')",
			Want:    "6",
		},
		{
			Formula: "=istext(sheet1!A1)",
			Want:    "true",
		},
		{
			Formula: "=istext(sheet1!B1)",
			Want:    "false",
		},
	}
	runTests(t, tests)
}

func testMathBuiltins(t *testing.T) {
	tests := []FormulaTestCase{
		{
			Formula: "=MIN(B1:B2)",
			Want:    "2",
		},
		{
			Formula: "=MAX(B1:B2)",
			Want:    "5",
		},
		{
			Formula: "=sum(sheet2!B1:B2, sheet1!B1:B2, 3) / 5",
			Want:    "5",
		},
		{
			Formula: "=pow(2, 2)",
			Want:    "4",
		},
		{
			Formula: "=abs(2)",
			Want:    "2",
		},
		{
			Formula: "=abs(-2)",
			Want:    "2",
		},
		{
			Formula: "=isnumber(sheet1!A1)",
			Want:    "false",
		},
		{
			Formula: "=isnumber(sheet1!B1)",
			Want:    "true",
		},
	}
	runTests(t, tests)
}

func runTests(t *testing.T, tests []FormulaTestCase) {
	t.Helper()
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
