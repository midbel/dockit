package eval

import (
	"bytes"
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
		t.Run("insert-row-simple", testInsertRowsBase)
		t.Run("insert-row-after", testInsertRowsAfter)
		t.Run("insert-row-before", testInsertRowsBefore)
	})
}

func testInsertRowsBase(t *testing.T) {

}

func testInsertRowsAfter(t *testing.T) {

}

func testInsertRowsBefore(t *testing.T) {

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
rows := @active.lines
cols := @active.columns
	`
	ev := runScript(t, script)
	checkValue(t, ev, "name", value.Text("sheet1"))
	checkValue(t, ev, "rows", value.Float(3))
	checkValue(t, ev, "cols", value.Float(3))

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
rows := @active.lines
cols := @active.columns
	`
	ev := runScript(t, script)
	checkValue(t, ev, "name", value.Text("sheet1"))
	checkValue(t, ev, "rows", value.Float(3))
	checkValue(t, ev, "cols", value.Float(3))

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
	if file, ok := view.(*types.File); ok {
		v, err := file.Active()
		if err != nil {
			t.Errorf("%s: active view not found", ident)
			return
		}
		view = v
	}
	v, ok := view.(*types.View)
	if !ok {
		t.Errorf("%s: expected view to be a View but got %T", ident, view)
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
	got, ok := view.(value.Array)
	if !ok {
		t.Errorf("%s: expected view to be an Array but got %T", ident, view)
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
	v, ok := view.(*types.View)
	if !ok {
		t.Errorf("%s: expected view to be a View but got %T", ident, view)
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
