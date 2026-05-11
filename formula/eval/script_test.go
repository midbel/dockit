package eval

import (
	"strings"
	"testing"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/value"
)

func TestScript(t *testing.T) {
	t.Run("basic", testBasicScript)
	t.Run("multiline", testMultilineBasicScript)
	t.Run("syntax-error", testSyntaxError)
	t.Run("undefined-identifier", testUndefinedIdentifier)
	t.Run("import-file", testImportFile)
	t.Run("slice-bounded-view", testSliceBoundedView)
}

func testSliceBoundedView(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `import "testdata/cities.csv" as cit
		use cit
		view := @active[A1:B3]`
		eg = createEngine()
	)
	_, err := eg.Exec(strings.NewReader(script), ev)
	if err != nil {
		t.Fatalf("slice bound failed due to unexpected error: %s", err)
	}
	view := ev.Resolve("view")
	if value.IsError(view) {
		t.Fatalf("view variable not defined: %s", view)
	}
	v, ok := view.(*types.View)
	if !ok {
		t.Fatalf("expected view to be types.View but got %T", view)
	}
	var (
		x = v.View()
		r = x.Bounds()
	)
	if r.Width() != 2 {
		t.Errorf("width mismatched! want 2, got %d", r.Width())
	}
	if r.Height() != 3 {
		t.Errorf("height mismatched! want 3, got %d", r.Height())
	}
}

func testImportFile(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `import "testdata/sample.csv" as data default
		import "testdata/countries.csv" using csv with tab as tab1 
		import "testdata/countries.csv" using csv as tab2
		foo := A1
		answer := B1`
		eg = createEngine()
	)
	_, err := eg.Exec(strings.NewReader(script), ev)
	if err != nil {
		t.Fatalf("import-file failed due to unexpected error: %s", err)
	}
	testEnv(t, ev, "foo", value.Text("foobar"))
	testEnv(t, ev, "answer", value.Float(42))
}

func testMultilineBasicScript(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `
		fourty2 := 42
		a := 1
		b := 1
		add := a + b
		sub := a - b
		div := a / b
		mul := a * b
		pow := a ^ b
		tpl := "answer is ${fourty2}"`
		eg = NewEngine()
	)
	_, err := eg.Exec(strings.NewReader(script), ev)
	if err != nil {
		t.Fatalf("basic script failed due to unexpected error: %s", err)
	}
	testEnv(t, ev, "add", value.Float(2))
	testEnv(t, ev, "sub", value.Float(0))
	testEnv(t, ev, "div", value.Float(1))
	testEnv(t, ev, "mul", value.Float(1))
	testEnv(t, ev, "pow", value.Float(1))
	testEnv(t, ev, "tpl", value.Text("answer is 42"))
}

func testBasicScript(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `
		name := upper(foo & 'bar') 
		answer`
		eg = createEngine()
	)
	ev.Define("foo", value.Text("foo"))
	ev.Define("answer", value.Float(42))
	got, err := eg.Exec(strings.NewReader(script), ev)
	if err != nil {
		t.Fatalf("basic script failed due to unexpected error: %s", err)
	}
	want := value.Float(42)
	if !isEqual(want, got) {
		t.Fatalf("result mismatched! want %v, got %v", want, got)
	}
	testEnv(t, ev, "name", value.Text("FOOBAR"))
}

func testSyntaxError(t *testing.T) {
	var (
		script = `name :=`
		eg     = createEngine()
	)

	_, err := eg.Exec(strings.NewReader(script), env.Empty())
	if err == nil {
		t.Fatal("syntax error expected but none returned")
	}
}

func testUndefinedIdentifier(t *testing.T) {
	var (
		script = `foo + missing`
		eg     = createEngine()
	)

	got, err := eg.Exec(strings.NewReader(script), env.Empty())
	if err != nil {
		t.Fatalf("basic script failed due to unexpected error: %s", err)
	}
	if !value.IsError(got) {
		t.Fatalf("errors expected, got %s", got)
	}
	if got != value.ErrRef {
		t.Fatalf("errors mismatched! want %s, got %s", value.ErrRef, got)
	}
}

func testEnv(t *testing.T, ev *env.Environment, ident string, want value.Value) {
	t.Helper()
	got := ev.Resolve(ident)
	if value.IsError(got) {
		t.Errorf("%s variable is error: %v", ident, got)
	}
	if !isEqual(got, want) {
		t.Errorf("%s value mismatched! want %s, got %s", ident, want, got)
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
