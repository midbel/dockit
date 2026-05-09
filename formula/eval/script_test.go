package eval

import (
	"strings"
	"testing"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/value"
)

func TestScript(t *testing.T) {
	t.Run("basic", testBasicScript)
	t.Run("multiline", testMultilineBasicScript)
	t.Run("syntax-error", testSyntaxError)
	t.Run("undefined-identifier", testUndefinedIdentifier)
	t.Run("import-file", testImportFile)
}

func testImportFile(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `import "testdata/sample.csv" as data default
		foo := A1
		answer := B1`
		eg = NewEngine()
	)
	eg.SetContextDir(".")
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
		a := 1
		b := 1
		add := a + b
		sub := a - b
		div := a / b
		mul := a * b
		pow := a ^ b`
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
}

func testBasicScript(t *testing.T) {
	var (
		ev     = env.Empty()
		script = `
		name := upper(foo & 'bar') 
		answer`
		eg = NewEngine()
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
		eg     = NewEngine()
	)

	_, err := eg.Exec(strings.NewReader(script), env.Empty())
	if err == nil {
		t.Fatal("syntax error expected but none returned")
	}
}

func testUndefinedIdentifier(t *testing.T) {
	var (
		script = `foo + missing`
		eg     = NewEngine()
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
