package grid_test

import (
	"fmt"
	"testing"

	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/layout"
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
			Comment: "division by zero",
			Formula: "=1/(1-1)",
			Want:    "#DIV/0!",
		},
		{
			Comment: "incompatible types",
			Formula: "=\"foo\"+1",
			Want:    "#VALUE!",
		},
		{
			Comment: "incompatible types",
			Formula: "=-\"foo\"+1",
			Want:    "#VALUE!",
		},
		{
			Comment: "incompatible types",
			Formula: "min(-\"test\", 1, 2)",
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

func TestFormula(t *testing.T) {
	t.Run("basic", testBasic)
	t.Run("basic-generated", testBasicGenerated)
	t.Run("compare", testCompare)
	t.Run("compare-generated", testCompareGenerated)
	t.Run("included-formula", testIncludedFormula)
}

func testIncludedFormula(t *testing.T) {
	tests := []FormulaTestCase{
		{
			Formula: "=C1",
			Want:    "24",
		},
		{
			Formula: "=C2",
			Want:    "20",
		},
		{
			Formula: "=sheet2!C1",
			Want:    "QUZ",
		},
		{
			Formula: "=sheet2!C2",
			Want:    "BEE",
		},
		{
			Formula: "=sheet2!D1",
			Want:    "FOO",
		},
		{
			Formula: "=sheet2!D2",
			Want:    "BAR",
		},
	}
	runTests(t, tests)
}

func testBasic(t *testing.T) {
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

func genTests(values, operators []string) []string {
	var tests []string
	for _, left := range values {
		for _, right := range values {
			for _, op := range operators {
				c := fmt.Sprintf("=%s %s %s", left, op, right)
				tests = append(tests, c)
			}
		}
	}
	return tests
}

func testBasicGenerated(t *testing.T) {
	values := []string{
		"1",
		"2",
		"\"foo\"",
		"\"bar\"",
		"true",
		"false",
	}

	operators := []string{"+", "-", "*", "/", "^", "&"}

	ctx := getContext()
	for _, c := range genTests(values, operators) {
		val, err := grid.EvalString(c, ctx)
		if err != nil {
			t.Errorf("%s: error executing formula: %s", c, err)
			continue
		}
		if !value.IsScalar(val) && !value.IsError(val) {
			t.Errorf("arithemetic should produces a scalar value")
			continue
		}
		assertKnownError(t, val)
	}
}

func testCompareGenerated(t *testing.T) {
	values := []string{
		"1",
		"5",
		"\"foo\"",
		"\"bar\"",
		"true",
		"false",
	}
	operators := []string{"=", "<>", ">", ">=", "<", "<="}

	ctx := getContext()
	for _, c := range genTests(values, operators) {
		val, err := grid.EvalString(c, ctx)
		if err != nil {
			t.Errorf("%s: error executing formula: %s", c, err)
			continue
		}
		if !value.IsScalar(val) && !value.IsError(val) {
			t.Errorf("comparison should produces a scalar value")
			continue
		}
		assertBoolResult(t, val)
		assertKnownError(t, val)
	}
}

func assertBoolResult(t *testing.T, val value.Value) {
	if value.IsScalar(val) {
		if got := val.String(); got != "true" && got != "false" {
			t.Errorf("boolean should be true or false! got %s", got)
		}
	}
}

func assertKnownError(t *testing.T, val value.Value) {
	if !value.IsError(val) {
		return
	}
	switch val {
	case value.ErrNull:
	case value.ErrDiv0:
	case value.ErrValue:
	case value.ErrRef:
	case value.ErrName:
	case value.ErrNum:
	case value.ErrNA:
	default:
		t.Errorf("unknown error return: %s", val.String())
	}
}

func testCompare(t *testing.T) {
	tests := []FormulaTestCase{
		{
			Formula: "=1=1",
			Want:    "true",
		},
		{
			Formula: "=1<>1",
			Want:    "false",
		},
		{
			Formula: "=1<>2",
			Want:    "true",
		},
		{
			Formula: "=1=2",
			Want:    "false",
		},
		{
			Formula: "=1<2",
			Want:    "true",
		},
		{
			Formula: "=1<=2",
			Want:    "true",
		},
		{
			Formula: "=2<=1",
			Want:    "false",
		},
		{
			Formula: "=1<=1",
			Want:    "true",
		},
		{
			Formula: "=1>=1",
			Want:    "true",
		},
		{
			Formula: "=1>=2",
			Want:    "false",
		},
		{
			Formula: "=2>=1",
			Want:    "true",
		},
		{
			Formula: "=\"foobar\" = \"foobar\"",
			Want:    "true",
		},
		{
			Formula: "=\"foo\" = \"bar\"",
			Want:    "false",
		},
		{
			Formula: "=\"foo\" <> \"bar\"",
			Want:    "true",
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
	f1, _ := grid.ParseFormula("=(B1 + sheet2!B1)*2")
	f2, _ := grid.ParseFormula("=(B2 + sheet2!B2)*2")
	sheet1.SetFormula(layout.NewPosition(1, 3), f1)
	sheet1.SetFormula(layout.NewPosition(2, 3), f2)

	sheet2 := flat.NewSheet("sheet2", value.Rows(
		[]value.ScalarValue{value.Text("quz"), value.Float(10)},
		[]value.ScalarValue{value.Text("bee"), value.Float(5)},
	))
	f3, _ := grid.ParseFormula("=UPPER(A1)")
	f4, _ := grid.ParseFormula("=UPPER(A2)")
	f5, _ := grid.ParseFormula("=UPPER(sheet1!A1)")
	f6, _ := grid.ParseFormula("=UPPER(sheet1!A2)")
	sheet2.SetFormula(layout.NewPosition(1, 3), f3)
	sheet2.SetFormula(layout.NewPosition(1, 4), f5)
	sheet2.SetFormula(layout.NewPosition(2, 3), f4)
	sheet2.SetFormula(layout.NewPosition(2, 4), f6)

	sheet3 := flat.NewSheet("sheet3", value.Rows(
		[]value.ScalarValue{value.Text("sum")},
		[]value.ScalarValue{value.Text("avg")},
		[]value.ScalarValue{value.Text("min")},
		[]value.ScalarValue{value.Text("max")},
	))

	file := flat.NewFileFromSheets(sheet1, sheet2, sheet3)

	return grid.NewContext(grid.FileContext(file))
}
