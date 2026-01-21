package formula

import (
	"testing"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type fakeContext struct {
	parent value.Context
	cells  map[string]value.Value
}

func (c fakeContext) At(pos layout.Position) (value.Value, error) {
	val, ok := c.cells[pos.String()]
	if !ok {
		return ErrValue, nil
	}
	return val, nil
}

func (c fakeContext) Range(start, end layout.Position) (value.Value, error) {
	return nil, nil
}

func (c fakeContext) Resolve(ident string) (value.Value, error) {
	return c.parent.Resolve(ident)
}

func one(_ []value.Value) (value.Value, error) {
	return Float(1), nil
}

func fake() value.Context {
	root := Empty()
	root.Define("one", NewFunction("one", one))

	ctx := fakeContext{
		parent: root,
		cells:  make(map[string]value.Value),
	}
	ctx.cells["A1"] = Float(0)
	ctx.cells["B1"] = Text("foo")
	ctx.cells["C1"] = Float(11)
	ctx.cells["A2"] = Float(0)
	ctx.cells["B2"] = Text("bar")
	ctx.cells["C2"] = Float(42)
	ctx.cells["A3"] = Float(0)
	ctx.cells["B3"] = Text("qux")
	ctx.cells["C3"] = Float(67)

	return ctx
}

func TestBasicFormula(t *testing.T) {
	ctx := fake()
	tests := []struct {
		Expr string
		Want string
	}{
		{
			Expr: "1+1",
			Want: "2",
		},
		{
			Expr: "'foo' & 'bar'",
			Want: "foobar",
		},
		{
			Expr: "one()",
			Want: "1",
		},
		{
			Expr: "one()/1",
			Want: "1",
		},
		{
			Expr: "$B$1",
			Want: "foo",
		},
		{
			Expr: "$B$1 & B2",
			Want: "foobar",
		},
		{
			Expr: "1 < 2",
			Want: "true",
		},
		{
			Expr: "1 = 2",
			Want: "false",
		},
		{
			Expr: "1 <> 2",
			Want: "true",
		},
		{
			Expr: "2 > 1",
			Want: "true",
		},
		{
			Expr: "1 >= 1",
			Want: "true",
		},
		{
			Expr: "1 >= 2",
			Want: "false",
		},
		{
			Expr: "'foo' = 'foo'",
			Want: "true",
		},
		{
			Expr: "true <> false",
			Want: "true",
		},
	}
	for _, c := range tests {
		expr, err := ParseFormula(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse formula: %s", c.Expr, err)
			continue
		}
		got, err := Eval(expr, ctx)
		if err != nil {
			t.Errorf("%s: error while evaluating formula: %s", c.Expr, err)
			continue
		}
		if got == nil {
			t.Errorf("%s: nil value", c.Expr)
			continue
		}
		if got.String() != c.Want {
			t.Errorf("%s: result mismatched! want %s, got %s", c.Expr, c.Want, got)
		}
	}
}

func TestFormulaAst(t *testing.T) {
	tests := []struct {
		Expr string
		Want string
	}{
		{
			Expr: "foobar",
			Want: "identifier(foobar)",
		},
		{
			Expr: "'foobar'",
			Want: "literal(foobar)",
		},
		{
			Expr: "42",
			Want: "number(42)",
		},
		{
			Expr: "-42",
			Want: "unary(number(42), -)",
		},
		{
			Expr: "'foo' & 'bar'",
			Want: "binary(literal(foo), literal(bar), &)",
		},
		{
			Expr: "1+1*2",
			Want: "binary(number(1), binary(number(1), number(2), *), +)",
		},
		{
			Expr: "one()",
			Want: "call(identifier(one), args: )",
		},
		{
			Expr: "test(1+1, 42)",
			Want: "call(identifier(test), args: binary(number(1), number(1), +), number(42))",
		},
		{
			Expr: "A2",
			Want: "cell(A2, false, false)",
		},
		{
			Expr: "$A2",
			Want: "cell(A2, true, false)",
		},
		{
			Expr: "$A$2",
			Want: "cell(A2, true, true)",
		},
		{
			Expr: "A1:A1000",
			Want: "range(cell(A1, false, false), cell(A1000, false, false))",
		},
	}
	for _, c := range tests {
		expr, err := ParseFormula(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expression: %s", c.Expr, err)
		}
		if got := DumpExpr(expr); got != c.Want {
			t.Errorf("%s: result mismatched! want %s, got %s", c.Expr, c.Want, got)
		}
	}
}
