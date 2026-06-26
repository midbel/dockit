package eval

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/value"
)

func TestScript(t *testing.T) {
	t.Run("values", func(t *testing.T) {
		t.Run("literals", testLiterals)
		t.Run("templates", testTemplates)
		t.Run("cells", testCellAccess)
		t.Run("array", testArrays)
		t.Run("ranges", testRanges)
		t.Run("slices-bounded", testSliceBounded)
		t.Run("slices-selection", testSliceSelection)
		t.Run("slices-filter", testSliceFilter)
	})
	t.Run("metadata", testMetadata)
	t.Run("errors", func(t *testing.T) {
		t.Run("syntax", testSyntaxError)
		t.Run("undefined-identifier", testUndefinedIdentifier)
	})
	t.Run("import-file", func(t *testing.T) {
		t.Run("json", testImportJson)
		t.Run("xml", testImportXml)
	})
	t.Run("export", testExport)
	t.Run("assert", func(t *testing.T) {
		t.Run("assertion-ok", testAssertOk)
		t.Run("assertion-fail", testAssertFail)
		t.Run("assertion-fail-ignore", testAssertFailIgnore)
		t.Run("assertion-fail-warning", testAssertFailWarning)
	})
	t.Run("print", testPrint)
	t.Run("use", testUse)
	t.Run("insert", func(t *testing.T) {
		t.Run("insert-rows", testInsertRows)
		t.Run("insert-columns", testInsertColumns)
	})
}

func testInsertRows(t *testing.T) {
	tests := []struct {
		Name   string
		Script string
		Cols   int64
		Rows   int64
		Want   [][]value.ScalarValue
	}{
		{
			Name: "row-basic",
			Script: `
import "testdata/salaries.csv" using csv[[comma]] as sh default
insert row into @active with 0
insrow := A4:C4`,
			Cols: 3,
			Rows: 4,
			Want: [][]value.ScalarValue{
				{value.Float(0), value.Float(0), value.Float(0)},
			},
		},
		{
			Name: "row-copy-line",
			Script: `
import "testdata/salaries.csv" using csv[[comma]] as sh default
insert row into @active with A1:C1
insrow := A4:C4`,
			Cols: 3,
			Rows: 4,
			Want: [][]value.ScalarValue{
				{value.Text("name"), value.Text("salary"), value.Text("bonus")},
			},
		},
		{
			Name: "row-before",
			Script: `
import "testdata/salaries.csv" using csv[[comma]] as sh default
insert row before 1 into @active
insrow := A1:C1`,
			Cols: 3,
			Rows: 4,
			Want: [][]value.ScalarValue{
				{value.Empty(), value.Empty(), value.Empty()},
			},
		},
		{
			Name: "row-after",
			Script: `
import "testdata/salaries.csv" using csv[[comma]] as sh default
insert 2 rows after 1 into @active
insrow := A2:C3`,
			Cols: 3,
			Rows: 5,
			Want: [][]value.ScalarValue{
				{value.Empty(), value.Empty(), value.Empty()},
				{value.Empty(), value.Empty(), value.Empty()},
			},
		},
		{
			Name: "row-multi",
			Script: `
import "testdata/salaries.csv" using csv[[comma]] as sh default
insert row before 1 into @active
insert row into @active
insrow := A1:C1`,
			Cols: 3,
			Rows: 5,
			Want: [][]value.ScalarValue{
				{value.Empty(), value.Empty(), value.Empty()},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ev := runScript(t, tt.Script)
			checkView(t, ev, "sh", tt.Cols, tt.Rows)
			checkRange(t, ev, "insrow", value.NewArray(tt.Want).(value.Array))
		})
	}
}

func testInsertColumns(t *testing.T) {
	t.SkipNow()
}

func testAssertOk(t *testing.T) {
	var (
		script = `assert 1 = 1`
		engine = createEngine()
	)
	_, err := engine.Exec(strings.NewReader(script), env.Empty())
	if err != nil {
		t.Errorf("expected assertion to pass! got error: %s", err)
	}
}

func testAssertFail(t *testing.T) {
	var (
		script = `assert 1 <> 1`
		engine = createEngine()
	)
	_, err := engine.Exec(strings.NewReader(script), env.Empty())
	if err == nil {
		t.Errorf("expected assertion to fail")
	}
	if _, ok := err.(AbortError); !ok {
		t.Errorf("expected error to be of type AbortError, got %T", err)
	}
}

func testAssertFailIgnore(t *testing.T) {
	var (
		script = `assert as ignore 1 <> 1`
		engine = createEngine()
	)
	_, err := engine.Exec(strings.NewReader(script), env.Empty())
	if err != nil {
		t.Errorf("expected assertion to fail without error, got %s", err)
	}
}

func testAssertFailWarning(t *testing.T) {
	var (
		script = `assert as warn 1 <> 1`
		engine = createEngine()
	)
	got, err := engine.Exec(strings.NewReader(script), env.Empty())
	if err != nil {
		t.Errorf("expected assertion to fail without error, got %s", err)
	}
	if got != value.ErrValue {
		t.Errorf("result mismatched! want %s, got %s", value.ErrValue, got)
	}
}

func testPrint(t *testing.T) {
	t.SkipNow()
}

func testUse(t *testing.T) {
	t.SkipNow()
}

func testExport(t *testing.T) {
	t.SkipNow()
}

func testImportJson(t *testing.T) {
	script := `
import "testdata/lang.json" using json[[$.owner.name, $.languages.name, $.languages.star | 0]] default

name := lang@active.name
rs := @active.lines
cs := @active.columns
	`
	ev := runScript(t, script)
	checkValue(t, ev, "name", value.Text("sheet1"))
	checkValue(t, ev, "rs", value.Float(3))
	checkValue(t, ev, "cs", value.Float(3))

	want := [][]value.ScalarValue{
		{value.Text("midbel"), value.Text("go"), value.Float(10)},
		{value.Text("midbel"), value.Text("rust"), value.Float(0)},
		{value.Text("midbel"), value.Text("python"), value.Float(6)},
	}
	checkArray(t, ev, "lang", value.NewArray(want).(value.Array))
}

func testImportXml(t *testing.T) {
	script := `
import "testdata/lang.xml" using xml[[$.owner.name, $.languages.language.name, $.languages.language.star:as("number") | 0]] default

name := lang@active.name
rs := @active.lines
cs := @active.columns
	`
	ev := runScript(t, script)
	checkValue(t, ev, "name", value.Text("sheet1"))
	checkValue(t, ev, "rs", value.Float(3))
	checkValue(t, ev, "cs", value.Float(3))

	want := [][]value.ScalarValue{
		{value.Text("midbel"), value.Text("go"), value.Float(10)},
		{value.Text("midbel"), value.Text("rust"), value.Float(6)},
		{value.Text("midbel"), value.Text("python"), value.Float(6)},
	}
	checkArray(t, ev, "lang", value.NewArray(want).(value.Array))
}

func testSyntaxError(t *testing.T) {
	script := `name :=`
	execScript(t, script, nil)
}

func testUndefinedIdentifier(t *testing.T) {
	var (
		script = `foo + missing`
		got    = execScript(t, script, nil)
	)
	if !value.IsError(got) {
		t.Fatalf("errors expected, got %s", got)
	}
	if got != value.ErrRef {
		t.Fatalf("errors mismatched! want %s, got %s", value.ErrRef, got)
	}
}

func testRanges(t *testing.T) {
	script := `
import "testdata/salaries.csv" using csv[[comma]] as dat default

salaries := B2:B3
bonus := C2:C3 * 2
raises := salaries + 1
totals := raises + bonus
	`
	ev := runScript(t, script)

	salaries := [][]value.ScalarValue{
		{value.Float(60)},
		{value.Float(50)},
	}
	checkRange(t, ev, "salaries", value.NewArray(salaries).(value.Array))
	bonus := [][]value.ScalarValue{
		{value.Float(10)},
		{value.Float(8)},
	}
	checkRange(t, ev, "bonus", value.NewArray(bonus).(value.Array))
	totals := [][]value.ScalarValue{
		{value.Float(71)},
		{value.Float(59)},
	}
	checkRange(t, ev, "totals", value.NewArray(totals).(value.Array))
}

func testArrays(t *testing.T) {
	script := `
import "testdata/salaries.csv" using csv[[comma]] as dat default

B2:B3 := 1 + B2:B3
C2:C3 := C2:C3 / 2
D2:D3 := B2:B3 + C2:C3
	`
	ev := runScript(t, script)

	want := [][]value.ScalarValue{
		{
			value.Text("name"),
			value.Text("salary"),
			value.Text("bonus"),
		},
		{
			value.Text("A"),
			value.Float(61),
			value.Float(2.5),
			value.Float(63.5),
		},
		{
			value.Text("B"),
			value.Float(51),
			value.Float(2),
			value.Float(53),
		},
	}
	checkArray(t, ev, "dat", value.NewArray(want).(value.Array))
}

func testSliceFilter(t *testing.T) {
	script := ``
	ev := runScript(t, script)
	_ = ev
	t.SkipNow()
}

func testSliceBounded(t *testing.T) {
	script := ``
	ev := runScript(t, script)
	_ = ev
	t.SkipNow()
}

func testSliceSelection(t *testing.T) {
	script := ``
	ev := runScript(t, script)
	_ = ev
	t.SkipNow()
}

func testLiterals(t *testing.T) {
	script := `
num := 42
str := "foobar"
truth := 42 >= 0	
	`
	ev := runScript(t, script)
	checkValue(t, ev, "num", value.Float(42))
	checkValue(t, ev, "str", value.Text("foobar"))
	checkValue(t, ev, "truth", value.Boolean(true))
}

func testCellAccess(t *testing.T) {
	script := `
import "testdata/repo.csv" using csv[[comma]] as repo default

stars := repo@active!B2 * 10
foobar := repo.sheet!A2 & "bar"
	`
	ev := runScript(t, script)
	checkValue(t, ev, "stars", value.Float(100))
	checkValue(t, ev, "foobar", value.Text("foobar"))
}

func testMetadata(t *testing.T) {
	script := `
import "testdata/repo.csv" using csv[[comma]] as repo default

sheet := @active.name
rows := @active.lines
cols := @active.columns
count := repo.sheets
	`
	ev := runScript(t, script)
	checkValue(t, ev, "sheet", value.Text("sheet"))
	checkValue(t, ev, "rows", value.Float(31))
	checkValue(t, ev, "cols", value.Float(7))
	checkValue(t, ev, "count", value.Float(1))
}

func testTemplates(t *testing.T) {
	script := `
import "testdata/repo.csv" using csv[[comma]] as repo default

# templates
template := "star of ${A2} = ${B2}"
	`
	ev := runScript(t, script)
	checkValue(t, ev, "template", value.Text("star of foo = 10"))
}

func checkValue(t *testing.T, ev *env.Environment, ident string, want value.Value) {
	t.Helper()
	got := ev.Resolve(ident)
	if value.IsError(got) {
		t.Errorf("%s: variable not defined", ident)
	}
	if !isEqual(got, want) {
		t.Errorf("%s: value mismatched! want %v, got %v", ident, want, got)
	}
}

func checkArray(t *testing.T, ev *env.Environment, ident string, want value.Array) {
	t.Helper()
	view := ev.Resolve(ident)
	if value.IsError(view) {
		t.Errorf("%s: view variable not defined", ident)
		return
	}
	v, err := getViewFromValue(view)
	if err != nil {
		t.Errorf("%s: %s", ident, err)
		return
	}
	got, ok := v.AsArray().(interface{ Equal(value.Array) bool })
	if !ok {
		t.Errorf("array are not comparable!")
		return
	}
	if !got.Equal(want) {
		t.Errorf("array mismatched! want %#v, got %#v", want, got)
	}
}

func checkRange(t *testing.T, ev *env.Environment, ident string, want value.Array) {
	t.Helper()
	view := ev.Resolve(ident)
	if value.IsError(view) {
		t.Errorf("%s: view variable not defined", ident)
		return
	}
	got, err := getArrayFromValue(view)
	if err != nil {
		t.Errorf("%s: %s", ident, err)
		return
	}
	if !got.Equal(want) {
		t.Errorf("array mismatched! want %#v, got %#v", want, got)
	}
}

func checkView(t *testing.T, ev *env.Environment, ident string, cols, rows int64) {
	t.Helper()
	view := ev.Resolve(ident)
	if value.IsError(view) {
		t.Errorf("%s: view variable not defined", ident)
		return
	}
	v, err := getViewFromValue(view)
	if err != nil {
		t.Errorf("%s: %s", ident, err)
		return
	}
	var (
		x = v.View()
		r = x.Bounds()
	)
	if r.Width() != cols {
		t.Errorf("columns number mismatched! want %d, got %d", cols, r.Width())
	}
	if r.Height() != rows {
		t.Errorf("rows number mismatched! want %d, got %d", rows, r.Height())
	}
}

func getArrayFromValue(arr value.Value) (value.Array, error) {
	if arr, ok := arr.(value.Array); ok {
		return arr, nil
	}
	if arr, ok := arr.(interface{ AsArray() value.ArrayValue }); ok {
		return getArrayFromValue(arr.AsArray())
	}
	var x value.Array
	return x, fmt.Errorf("expected value to be an Array, got %T", arr)
}

func getViewFromValue(view value.Value) (*types.View, error) {
	switch x := view.(type) {
	case *types.File:
		val, err := x.Active()
		if err != nil {
			return nil, fmt.Errorf("file has no active sheet")
		}
		return getViewFromValue(val)
	case *types.View:
		return x, nil
	default:
		return nil, fmt.Errorf("expected value to be a view/file but got %T", view)
	}
}

func isEqual(got, want value.Value) bool {
	ok := value.Eq(got, want)
	return value.True(ok)
}

func createEngine() *Engine {
	eg := NewEngine()
	eg.SetContextDir(".")
	return eg
}

func runScript(t *testing.T, script string) *env.Environment {
	t.Helper()

	ev := env.Empty()
	execScript(t, script, ev)
	return ev
}

func execScript(t *testing.T, script string, ev *env.Environment) value.Value {
	t.Helper()

	eg := createEngine()
	eg.Stdout = bytes.NewBuffer(nil)
	eg.Stderr = bytes.NewBuffer(nil)
	if ev == nil {
		ev = env.Empty()
	}
	got, err := eg.Exec(strings.NewReader(script), ev)
	if err != nil {
		t.Fatalf("error executing script: %s", err)
	}
	return got
}
