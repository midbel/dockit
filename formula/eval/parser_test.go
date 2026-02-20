package eval

import (
	"testing"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/layout"
)

type importExpect struct {
	File      string
	Format    string
	Specifier string
	Alias     string
	Default   bool
	Readonly  bool
}

func TestImportStmt(t *testing.T) {
	tests := []struct {
		Expr   string
		Expect importExpect
	}{
		{
			Expr: "import \"file.csv\"",
			Expect: importExpect{
				File: "file.csv",
			},
		},
		{
			Expr: "import \"file.csv\" default",
			Expect: importExpect{
				File:    "file.csv",
				Default: true,
			},
		},
		{
			Expr: "import \"file.csv\" ro",
			Expect: importExpect{
				File:     "file.csv",
				Readonly: true,
			},
		},
		{
			Expr: "import \"file.csv\" default rw",
			Expect: importExpect{
				File:     "file.csv",
				Default:  true,
				Readonly: false,
			},
		},
		{
			Expr: "import \"file.csv\" using csv as file",
			Expect: importExpect{
				File:   "file.csv",
				Alias:  "file",
				Format: "csv",
			},
		},
	}
	p := NewParser(ScriptGrammar())
	for _, c := range tests {
		expr, err := p.ParseString(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expr: %s", c.Expr, err)
			continue
		}
		got, ok := unwrapScriptExpr(expr).(importFile)
		if !ok {
			t.Errorf("%s: expected import statement, got %T", c.Expr, expr)
			continue
		}
		assertImportRef(t, c.Expr, got, c.Expect)
	}
}

func assertImportRef(t *testing.T, expr string, got importFile, want importExpect) {
	t.Helper()
	if want.File != got.file {
		t.Errorf("%s: file mismatched! want %s, got %s", expr, want.File, got.file)
	}
	if want.Alias != got.alias {
		t.Errorf("%s: alias mismatched! want %s, got %s", expr, want.Alias, got.alias)
	}
	if want.Format != got.format {
		t.Errorf("%s: format mismatched! want %s, got %s", expr, want.Format, got.format)
	}
	if want.Default != got.defaultFile {
		t.Errorf("%s: default mismatched! want %t, got %t", expr, want.Default, got.defaultFile)
	}
	if want.Readonly != got.readOnly {
		t.Errorf("%s: readonly mismatched! want %t, got %t", expr, want.Readonly, got.readOnly)
	}
}

type useExpect struct {
	Value    string
	Readonly bool
}

func TestUseStmt(t *testing.T) {
	tests := []struct {
		Expr   string
		Expect useExpect
	}{
		{
			Expr: "use view",
			Expect: useExpect{
				Value: "view",
			},
		},
		{
			Expr: "use view ro",
			Expect: useExpect{
				Value:    "view",
				Readonly: true,
			},
		},
		{
			Expr: "use view rw",
			Expect: useExpect{
				Value:    "view",
				Readonly: false,
			},
		},
	}
	p := NewParser(ScriptGrammar())
	for _, c := range tests {
		expr, err := p.ParseString(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expr: %s", c.Expr, err)
			continue
		}
		got, ok := unwrapScriptExpr(expr).(useRef)
		if !ok {
			t.Errorf("%s: expected use statement, got %T", c.Expr, expr)
			continue
		}
		assertUseRef(t, c.Expr, got, c.Expect)
	}
}

func assertUseRef(t *testing.T, expr string, got useRef, want useExpect) {
	t.Helper()
	if want.Value != got.ident {
		t.Errorf("%s: identifier mismatched! want %s, got %s", expr, want.Value, got.ident)
	}
	if want.Readonly != got.readOnly {
		t.Errorf("%s: readonly mismatched! want %t, got %t", expr, want.Readonly, got.readOnly)
	}
}

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

func TestExpr(t *testing.T) {
	tests := []struct {
		Expr string
		Want Expr
	}{
		{
			Expr: "x * 5 + 10",
			Want: NewBinary(
				NewBinary(
					NewIdentifier("x"),
					NewNumber(5),
					op.Mul,
				),
				NewNumber(10),
				op.Add,
			),
		},
		{
			Expr: "(x * 5) + 10",
			Want: NewBinary(
				NewBinary(
					NewIdentifier("x"),
					NewNumber(5),
					op.Mul,
				),
				NewNumber(10),
				op.Add,
			),
		},
		{
			Expr: "5 * (x + 10)",
			Want: NewBinary(
				NewNumber(5),
				NewBinary(
					NewIdentifier("x"),
					NewNumber(10),
					op.Add,
				),
				op.Mul,
			),
		},
		{
			Expr: "-5 * (x + 10)",
			Want: NewBinary(
				NewUnary(
					NewNumber(5),
					op.Sub,
				),
				NewBinary(
					NewIdentifier("x"),
					NewNumber(10),
					op.Add,
				),
				op.Mul,
			),
		},
		{
			Expr: "'foo' & 'bar'",
			Want: NewBinary(
				NewLiteral("foo"),
				NewLiteral("bar"),
				op.Concat,
			),
		},
		{
			Expr: "sum(A1:A100) + 1",
			Want: NewBinary(
				NewCall(
					NewIdentifier("sum"),
					[]Expr{
						NewRangeAddr(
							NewCellAddr(layout.NewPosition(1, 1), false, false),
							NewCellAddr(layout.NewPosition(100, 1), false, false),
						),
					},
				),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "sum($A1, A$100, $B$1) / 100",
			Want: NewBinary(
				NewCall(
					NewIdentifier("sum"),
					[]Expr{
						NewCellAddr(layout.NewPosition(1, 1), true, false),
						NewCellAddr(layout.NewPosition(100, 1), false, true),
						NewCellAddr(layout.NewPosition(1, 2), true, true),
					},
				),
				NewNumber(100),
				op.Div,
			),
		},
		{
			Expr: "A1 := min(A1, A100) + view!A1:A100",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewBinary(
					NewCall(
						NewIdentifier("min"),
						[]Expr{
							NewCellAddr(layout.NewPosition(1, 1), false, false),
							NewCellAddr(layout.NewPosition(100, 1), false, false),
						},
					),
					NewQualifiedAddr(
						NewIdentifier("view"),
						NewRangeAddr(
							NewCellAddr(layout.NewPosition(1, 1), false, false),
							NewCellAddr(layout.NewPosition(100, 1), false, false),
						),
					),
					op.Add,
				),
			),
		},
		{
			Expr: "A1 += 100",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewNumber(100),
					op.Add,
				),
			),
		},
		{
			Expr: "A1 *= age",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewIdentifier("age"),
					op.Mul,
				),
			),
		},
		{
			Expr: "A1 -= total",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewIdentifier("total"),
					op.Sub,
				),
			),
		},
		{
			Expr: "A1 /= 0",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewNumber(0),
					op.Div,
				),
			),
		},
		{
			Expr: "A1 &= 'foobar'",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(1, 1), false, false),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewLiteral("foobar"),
					op.Concat,
				),
			),
		},
	}
	p := NewParser(ScriptGrammar())
	for _, c := range tests {
		expr, err := p.ParseString(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expr: %s", c.Expr, err)
			continue
		}
		assertEqualExpr(t, c.Want, unwrapScriptExpr(expr))
	}
}

func assertEqualExpr(t *testing.T, want, got Expr) {
	t.Helper()
	switch w := want.(type) {
	case rangeAddr:
		g, ok := got.(rangeAddr)
		if !ok {
			t.Errorf("rangeAddr expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.startAddr, g.startAddr)
		assertEqualExpr(t, w.endAddr, g.endAddr)
	case cellAddr:
		g, ok := got.(cellAddr)
		if !ok {
			t.Errorf("cellAddr expression expected but got %T", got)
			return
		}
		if !w.Position.Equal(g.Position) {
			t.Errorf("position mismatched! want %s, got %s", w.Position, g.Position)
		}
		if w.AbsCol != g.AbsCol {
			t.Errorf("absolute column mismatched!")
		}
		if w.AbsRow != g.AbsRow {
			t.Errorf("absolute column mismatched!")
		}
	case qualifiedCellAddr:
		g, ok := got.(qualifiedCellAddr)
		if !ok {
			t.Errorf("qualifiedCellAddr expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.path, g.path)
		assertEqualExpr(t, w.addr, g.addr)
	case assignment:
		g, ok := got.(assignment)
		if !ok {
			t.Errorf("assignment expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.ident, g.ident)
		assertEqualExpr(t, w.expr, g.expr)
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
	case unary:
		g, ok := got.(unary)
		if !ok {
			t.Errorf("unary expression expected but got %T", got)
			return
		}
		if g.op != w.op {
			t.Errorf("unary operator mismatched!")
		}
		assertEqualExpr(t, w.expr, g.expr)
	case call:
		g, ok := got.(call)
		if !ok {
			t.Errorf("call expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.ident, g.ident)
		if len(w.args) != len(g.args) {
			t.Errorf("arguments count mismatched! want %d, got %d", len(w.args), len(g.args))
		}
		for i := range w.args {
			assertEqualExpr(t, w.args[i], g.args[i])
		}
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
