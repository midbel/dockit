package eval

import (
	"testing"

	"github.com/midbel/dockit/formula/op"
)

func TestPrintStmt(t *testing.T) {
	tests := []struct {
		Expr    string
		Value   Expr
		Pattern string
	}{
		{
			Expr:  "print x",
			Value: NewIdentifier("x"),
		},
		{
			Expr:  "print 'foobar'",
			Value: NewLiteral("foobar"),
		},
		{
			Expr: "print \"foo is ${foo}, age is ${42+0}\"",
			Value: NewTemplate([]Expr{
				NewLiteral("foo is "),
				NewIdentifier("foo"),
				NewLiteral(", age is "),
				NewBinary(NewNumber(42), NewNumber(0), op.Add),
			}),
		},
		{
			Expr: "print \"foo is ${foo}, age is ${42+0}!\"",
			Value: NewTemplate([]Expr{
				NewLiteral("foo is "),
				NewIdentifier("foo"),
				NewLiteral(", age is "),
				NewBinary(NewNumber(42), NewNumber(0), op.Add),
				NewLiteral("!"),
			}),
		},
		{
			Expr:    "print 3.14 '###.###'",
			Value:   NewNumber(3.14),
			Pattern: "###.###",
		},
		{
			Expr:  "print x * y",
			Value: NewBinary(NewIdentifier("x"), NewIdentifier("y"), op.Mul),
		},
	}
	p := NewParser(ScriptGrammar())
	for _, c := range tests {
		expr, err := p.ParseString(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expr: %s", c.Expr, err)
			continue
		}
		pr, ok := unwrapScriptExpr(expr).(printRef)
		if !ok {
			t.Errorf("%s: expected print statement, got %T", c.Expr, expr)
			continue
		}
		assertEqualExpr(t, c.Value, pr.expr)
		if c.Pattern != "" && c.Pattern != pr.pattern {
			t.Errorf("%s: pattern format mismatched! want %s, got %s", c.Expr, c.Pattern, pr.pattern)
		}
	}
}

func assertEqualExpr(t *testing.T, want, got Expr) {
	t.Helper()
	switch w := want.(type) {
	case binary:
		g, ok := got.(binary)
		if !ok {
			t.Errorf("binary expression expected but got %T", got)
			return
		}
		if g.op != w.op {
			t.Errorf("binary operator mismatched!")
		}
		assertEqualExpr(t, w.left, g.left)
		assertEqualExpr(t, w.right, g.right)
	case number:
		g, ok := got.(number)
		if !ok {
			t.Errorf("number expected but got %T", got)
			return
		}
		if w.value != g.value {
			t.Errorf("number value mismatched! want %f, got %f", w.value, g.value)
		}
	case literal:
		g, ok := got.(literal)
		if !ok {
			t.Errorf("literal expected but got %T", got)
			return
		}
		if w.value != g.value {
			t.Errorf("literal value mismatched! want %s, got %s", w.value, g.value)
		}
	case template:
		g, ok := got.(template)
		if !ok {
			t.Errorf("template expected but got %T", got)
			return
		}
		if len(g.expr) != len(w.expr) {
			t.Errorf("nested expressions count mismatched! want %d, got %d", len(w.expr), len(g.expr))
			return
		}
		for i := range w.expr {
			assertEqualExpr(t, w.expr[i], g.expr[i])
		}
	case identifier:
		g, ok := got.(identifier)
		if !ok {
			t.Errorf("identifier expected but got %T", got)
			return
		}
		if w.name != g.name {
			t.Errorf("identifier name mismatched! want %s, got %s", w.name, g.name)
		}
	default:
		t.Errorf("unsupported expression type %T", want)
	}
}

func unwrapScriptExpr(expr Expr) Expr {
	s, ok := expr.(Script)
	if ok && len(s.Body) == 1 {
		return s.Body[0]
	}
	return s
}
