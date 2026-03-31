package grid_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/testutil"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type FormulaTestCase struct {
	Formula string
	Want    string
	Sheet   string
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
			Formula: "=sum(1, 1/0)",
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
	ctx := testutil.FileContext()
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
			Formula: "=C1",
			Want:    "QUZ",
			Sheet:   "sheet2",
		},
		{
			Formula: "=C2",
			Want:    "BEE",
			Sheet:   "sheet2",
		},
		{
			Formula: "=D1",
			Want:    "FOO",
			Sheet:   "sheet2",
		},
		{
			Formula: "=D2",
			Want:    "BAR",
			Sheet:   "sheet2",
		},
		{
			Formula: "=B1",
			Want:    "7",
			Sheet:   "sheet3",
		},
		{
			Formula: "=B2",
			Want:    "3.5",
			Sheet:   "sheet3",
		},
		{
			Formula: "=B3",
			Want:    "5",
			Sheet:   "sheet3",
		},
		{
			Formula: "=B4",
			Want:    "10",
			Sheet:   "sheet3",
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

	ctx := testutil.FileContext()
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

	ctx := testutil.FileContext()
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

func TestSync(t *testing.T) {
	t.Run("force-sync", testForceSync)
	t.Run("empty-no-sync", testEmptyWithoutSync)
}

func testEmptyWithoutSync(t *testing.T) {
	var (
		file  = testutil.CreateFile()
		blank = value.Empty()
		pos   = layout.NewPosition(1, 3)
		want  = value.Float(24)
	)

	sh, err := file.Sheet("sheet1")
	if err != nil {
		t.Fatalf("fail to get sheet1: %s", err)
	}
	cell, err := sh.Cell(pos)
	if err != nil {
		t.Fatalf("fail to get cell at %s: %s", pos, err)
	}
	if got := cell.Value(); value.True(value.Ne(blank, got)) {
		t.Errorf("expected empty value in sheet1!C1! got %s", got)
	}
	if f := cell.Formula(); f == nil {
		t.Fatalf("expected formula to not be nil")
	} else {
		val, err := grid.Eval(f, grid.NewContext(grid.FileContext(file)))
		if err != nil {
			t.Fatalf("error evaluating formula: %s", err)
		}
		if value.True(value.Ne(val, want)) {
			t.Errorf("evaluation failed! want %s, got %s", want, val)
		}
	}
	if got := cell.Value(); value.True(value.Ne(blank, got)) {
		t.Errorf("expected empty value in sheet1!C1! got %s", got)
	}
}

const sample = `project,star,commit,language
foo,10,2023,Go
bar,13,452,C
flim,156,892,Rust
glam,42,1105,TypeScript
zorp,804,342,Go
munt,424,11127,C`

func TestViews(t *testing.T) {
	t.Run("bounded-view", testBoundedView)
	t.Run("project-view", testProjectView)
	t.Run("transpose-view", testTransposeView)
	t.Run("horizontal-stack-view", testHorizontalStackView)
	t.Run("vertical-stack-view", testVerticalStackView)
	t.Run("combined-view", testCombinedViews)
}

func testCombinedViews(t *testing.T) {

}

func testBoundedView(t *testing.T) {
	var (
		sheet = getSheetFromSample(t)
		sbd   = sheet.Bounds()
		rg    = layout.NewRange(
			layout.NewPosition(2, 1),
			layout.NewPosition(4, 2),
		)
		view = grid.NewBoundedView(sheet, rg)
		vbd  = view.Bounds()
	)
	if vbd.Width() != rg.Width() || vbd.Height() != rg.Height() {
		t.Fatalf("view bounds does not match building range")
	}
	if vbd.Width() == sbd.Width() && vbd.Height() == sbd.Height() {
		t.Fatalf("view bounds should not match sheet bounds")
	}

	for pos := range vbd.Positions() {
		other := pos.Offset(1, 0)

		var (
			cell1, _ = view.Cell(pos)
			cell2, _ = sheet.Cell(other)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", pos, other, cell1.Value(), cell2.Value())
		}
	}
}

func testProjectView(t *testing.T) {
	var (
		sheet   = getSheetFromSample(t)
		sbd     = sheet.Bounds()
		cols, _ = layout.SelectionFromString("A;D")
		view    = grid.NewProjectView(sheet, cols)
		vbd     = view.Bounds()
	)
	if vbd.Width() == sbd.Width() {
		t.Fatalf("view width should not match sheet width")
	}
	if vbd.Height() != sbd.Height() {
		t.Fatalf("view height should match sheet height")
	}
	var (
		other   layout.Position
		columns = []int64{1, 4}
	)
	for pos := range vbd.Positions() {
		other.Line = pos.Line
		other.Column = columns[pos.Column-1]
		var (
			cell1, _ = view.Cell(pos)
			cell2, _ = sheet.Cell(other)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", pos, other, cell1.Value(), cell2.Value())
		}
	}
}

func testTransposeView(t *testing.T) {
	var (
		sheet = getSheetFromSample(t)
		sbd   = sheet.Bounds()
		view  = grid.NewTransposedView(sheet)
		vbd   = view.Bounds()
	)
	if vbd.Width() != sbd.Height() {
		t.Fatalf("view width should be equal to sheet height")
	}
	if vbd.Height() != sbd.Width() {
		t.Fatalf("view height should be equal to sheet width")
	}
	var other layout.Position
	for pos := range vbd.Positions() {
		other.Line = pos.Column
		other.Column = pos.Line

		var (
			cell1, _ = view.Cell(pos)
			cell2, _ = sheet.Cell(other)
			ok       = value.Eq(cell1.Value(), cell2.Value())
		)
		if !value.True(ok) {
			t.Errorf("value mismatched at %s vs %s! want %s, got %s", pos, other, cell1.Value(), cell2.Value())
		}
	}
}

func testHorizontalStackView(t *testing.T) {

}

func testVerticalStackView(t *testing.T) {

}

func getSheetFromSample(t *testing.T) grid.View {
	t.Helper()

	file, err := testutil.CreateCsvFile(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("fail to create csv file: %s", err)
	}

	sheet, err := file.ActiveSheet()
	if err != nil {
		t.Fatalf("fail to retrieve active sheet: %s", err)
	}
	return sheet
}

func testForceSync(t *testing.T) {
	var (
		file  = testutil.CreateFile()
		blank = value.Empty()
		pos   = layout.NewPosition(1, 3)
		want  = value.Float(24)
	)

	sh, err := file.Sheet("sheet1")
	if err != nil {
		t.Fatalf("fail to get sheet1: %s", err)
	}
	cell, err := sh.Cell(pos)
	if err != nil {
		t.Fatalf("fail to get cell at %s: %s", pos, err)
	}
	if got := cell.Value(); value.True(value.Ne(blank, got)) {
		t.Errorf("expected empty value in sheet1!C1! got %s", got)
	}
	if err := file.Sync(); err != nil {
		t.Errorf("fail to sync file: %s", err)
	}
	if got := cell.Value(); value.True(value.Ne(want, got)) {
		t.Errorf("expected value in sheet1!C1 to be %s! got %s", want, got)
	}
}

func runTests(t *testing.T, tests []FormulaTestCase) {
	t.Helper()
	var (
		file = testutil.CreateFile()
		ctx  = grid.NewContext(grid.FileContext(file))
	)
	for _, c := range tests {
		sub := ctx
		if c.Sheet != "" {
			sh, err := file.Sheet(c.Sheet)
			if err != nil {
				t.Errorf("%s: sheet not found", c.Sheet)
				continue
			}
			sub = grid.EnclosedContext(sub, grid.SheetContext(sh))
		}
		val, err := grid.EvalString(c.Formula, sub)
		if err != nil {
			t.Errorf("%s: error executing formula: %s", c.Formula, err)
			continue
		}
		if got := val.String(); got != c.Want {
			t.Errorf("%s: result mismatched! want %s, got %s", c.Formula, c.Want, got)
		}
	}
}
