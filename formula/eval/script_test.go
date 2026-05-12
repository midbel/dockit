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
	t.Run("basic", testBasicScript)
	t.Run("multiline", testMultilineBasicScript)
	t.Run("syntax-error", testSyntaxError)
	t.Run("undefined-identifier", testUndefinedIdentifier)
	t.Run("import-file", testImportFile)
	t.Run("import-config-file", testImportConfigFile)
	t.Run("slices-view", testSlicesView)
}

func testSlicesView(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `import "testdata/cities.csv" as cit
		use cit
		bs := @active[A1:B3]
		cs := @active[A:B]
		fs := @active[C1 = "1"]`
	)
	execScript(t, script, ev)
	testView(t, ev, "bs", 2, 3)
	testView(t, ev, "cs", 2, 13)
	testView(t, ev, "fs", 4, 5)
}

func testImportConfigFile(t *testing.T) {
	script := `#!script
	#! csv.delimiter := tab
	#! log.pattern := "%t %l%b[%p]%b[%u:%g]%b%n:%b%m"

	import "testdata/countries.csv" default
	import "testdata/app.log" using log as app

	abbr := A1
	name := lower(B1)
	user := app.sheet!D1
	group := app.sheet!E1
	perm := user & ':' & group`
	ev := env.Empty()
	execScript(t, script, ev)
	testEnv(t, ev, "abbr", value.Text("be"))
	testEnv(t, ev, "name", value.Text("belgium"))
	testEnv(t, ev, "user", value.Text("alice"))
	testEnv(t, ev, "group", value.Text("admin"))
	testEnv(t, ev, "perm", value.Text("alice:admin"))
}

func testImportFile(t *testing.T) {
	script := `import "testdata/sample.csv" as data default
	import "testdata/countries.csv" using csv with tab as tab1 
	import "testdata/countries.csv" using csv as tab2
	foo := A1
	answer := B1`
	ev := env.Empty()
	execScript(t, script, ev)
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
	)
	execScript(t, script, ev)
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
	)
	ev.Define("foo", value.Text("foo"))
	ev.Define("answer", value.Float(42))
	var (
		got  = execScript(t, script, ev)
		want = value.Float(42)
	)
	if !isEqual(want, got) {
		t.Fatalf("result mismatched! want %v, got %v", want, got)
	}
	testEnv(t, ev, "name", value.Text("FOOBAR"))
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

func testView(t *testing.T, ev *env.Environment, ident string, cols, rows int64) {
	t.Helper()
	view := ev.Resolve(ident)
	if value.IsError(view) {
		t.Errorf("%s: view variable not defined", ident)
	}
	v, ok := view.(*types.View)
	if !ok {
		t.Errorf("%s: expected view to be a View but got %T", ident, view)
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
