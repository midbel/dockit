package formula

import (
	"testing"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

type fakeContext struct {
	parent value.Context
}

func (c fakeContext) At(pos layout.Position) (value.Value, error) {
	return nil, nil
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
	}
	return ctx
}

func TestBasic(t *testing.T) {
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
		if got.String() != c.Want {
			t.Errorf("%s: result mismatched! want %s, got %s", c.Expr, c.Want, got)
		}
	}
}
