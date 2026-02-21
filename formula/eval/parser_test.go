package eval

import (
	"fmt"
	"testing"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/layout"
)

func TestParseFormula(t *testing.T) {
	tests := []struct{
		Expr string
		Want Expr
	}{
		{
			Expr: "=1+1",
			Want: NewBinary(
				NewNumber(1),
				NewNumber(1),
				op.Add,
			),
		},
		{
			Expr: "=sum(A1, A2, A3)",
			Want: NewCall(
				NewIdentifier("sum"),
				[]Expr{
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewCellAddr(layout.NewPosition(2, 1), false, false),
					NewCellAddr(layout.NewPosition(3, 1), false, false),
				},
			),
		},
	}
	for _, c := range tests {
		f, err := ParseFormula(c.Expr)
		if err != nil {
			t.Errorf("%s: error parsing formumla: %s", c.Expr, err)
			continue
		}
		e, ok := f.(deferredFormula)
		if !ok {
			t.Errorf("formula expected - got %T", f)
			continue
		}
		assertEqualExpr(t, c.Want, e.expr)
	}
}

func TestSaveStmt(t *testing.T) {
	t.SkipNow()
}

func TestExportStmt(t *testing.T) {
	t.SkipNow()
}

func TestPushStmt(t *testing.T) {
	t.SkipNow()
}

func TestPopStmt(t *testing.T) {
	t.SkipNow()
}

func TestLockStmt(t *testing.T) {
	t.SkipNow()
}

func TestUnlockStmt(t *testing.T) {
	t.SkipNow()
}

func TestClearStmt(t *testing.T) {
	t.SkipNow()
}

type importExpect struct {
	File      string
	Format    string
	Specifier string
	Alias     string
	Default   bool
	Readonly  bool
	Options   map[string]string
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
		{
			Expr: "import \"file.log\" using log with \"%time %user %level %message\" as file",
			Expect: importExpect{
				File:      "file.log",
				Alias:     "file",
				Format:    "log",
				Specifier: "%time %user %level %message",
			},
		},
		{
			Expr: "import \"file.csv\" using csv with (quote := 'true', separator := 'tab') default ro",
			Expect: importExpect{
				File:     "file.csv",
				Format:   "csv",
				Default:  true,
				Readonly: true,
				Options: map[string]string{
					"quote":     "true",
					"separator": "tab",
				},
			},
		},
		{
			Expr: "import \"file.csv\" using csv with (\nquote := 'true',\nseparator := 'tab'\n) default ro",
			Expect: importExpect{
				File:     "file.csv",
				Format:   "csv",
				Default:  true,
				Readonly: true,
				Options: map[string]string{
					"quote":     "true",
					"separator": "tab",
				},
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
	if len(want.Options) != len(got.options) {
		t.Errorf("number of options mismatched! want %d, got %d", len(want.Options), len(got.options))
	}
	for k, v := range want.Options {
		other, ok := got.options[k]
		if !ok {
			t.Errorf("option %s not set", k)
			continue
		}
		if v != other {
			t.Errorf("value of option %s mismatched! want %s, got %s", k, v, other)
		}
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
			Expr: "file.sheets",
			Want: NewAccess(NewIdentifier("file"), "sheets"),
		},
		{
			Expr: "100 ^ 1%",
			Want: NewBinary(
				NewNumber(100),
				NewPostfix(NewNumber(1), op.Percent),
				op.Pow,
			),
		},
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
			Expr: "sum(\n$A1,\n A$100,\n $B$1\n) / 100",
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
		{
			Expr: "lambda := =sum(A1, B2)",
			Want: NewAssignment(
				NewIdentifier("lambda"),
				NewDeferred(NewCall(
					NewIdentifier("sum"),
					[]Expr{
						NewCellAddr(layout.NewPosition(1, 1), false, false),
						NewCellAddr(layout.NewPosition(2, 2), false, false),
					},
				)),
			),
		},
		{
			Expr: "E5 := =10 + 100 + view[A1:C100]",
			Want: NewAssignment(
				NewCellAddr(layout.NewPosition(5, 5), false, false),
				NewDeferred(
					NewBinary(
						NewBinary(
							NewNumber(10),
							NewNumber(100),
							op.Add,
						),
						NewSlice(
							NewIdentifier("view"),
							NewRangeAddr(
								NewCellAddr(layout.NewPosition(1, 1), false, false),
								NewCellAddr(layout.NewPosition(100, 3), false, false),
							),
						),
						op.Add,
					),
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

func TestSlices(t *testing.T) {
	tests := []struct {
		Expr string
		Want Expr
	}{
		{
			Expr: "view1[A;B;C]",
			Want: NewSlice(
				NewIdentifier("view1"),
				NewColumnsSlice([]Expr{
					SelectRange(1, 1, 0),
					SelectRange(2, 2, 0),
					SelectRange(3, 3, 0),
				}),
			),
		},
		{
			Expr: "view2[:E;B:D:2;C::3]",
			Want: NewSlice(
				NewIdentifier("view2"),
				NewColumnsSlice([]Expr{
					SelectRange(0, 5, 0),
					SelectRange(2, 4, 2),
					SelectRange(3, 0, 3),
				}),
			),
		},
		{
			Expr: "view3[A1:C2]",
			Want: NewSlice(
				NewIdentifier("view3"),
				NewRangeAddr(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewCellAddr(layout.NewPosition(2, 3), false, false),
				),
			),
		},
		{
			Expr: "view4[A:C]",
			Want: NewSlice(
				NewIdentifier("view4"),
				NewColumnsSlice([]Expr{
					SelectRange(1, 3, 0),
				}),
			),
		},
		{
			Expr: "view5[D1 >= 100 and A1 <> 'test']",
			Want: NewSlice(
				NewIdentifier("view5"),
				NewAnd(
					NewBinary(
						NewCellAddr(layout.NewPosition(1, 4), false, false),
						NewNumber(100),
						op.Ge,
					),
					NewBinary(
						NewCellAddr(layout.NewPosition(1, 1), false, false),
						NewLiteral("test"),
						op.Ne,
					),
				),
			),
		},
		{
			Expr: "view6[not D1 = 'foobar']",
			Want: NewSlice(
				NewIdentifier("view6"),
				NewNot(
					NewBinary(
						NewCellAddr(layout.NewPosition(1, 4), false, false),
						NewLiteral("foobar"),
						op.Eq,
					),
				),
			),
		},
		{
			Expr: "view7[D1 = 'foobar']",
			Want: NewSlice(
				NewIdentifier("view7"),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 4), false, false),
					NewLiteral("foobar"),
					op.Eq,
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
		got, ok := unwrapScriptExpr(expr).(slice)
		if !ok {
			t.Errorf("%s: expected slice expression, got %T", c.Expr, expr)
			continue
		}
		assertEqualExpr(t, c.Want, got)
	}
}

func TestScript(t *testing.T) {
	tests := []struct {
		Expr string
		Want Expr
	}{
		{
			Expr: "A1 := 100;; A1 * 2;",
			Want: NewScript([]Expr{
				NewAssignment(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewNumber(100),
				),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewNumber(2),
					op.Mul,
				),
			}),
		},
		{
			Expr: "A1 := 100;;\n A1 * 2;",
			Want: NewScript([]Expr{
				NewAssignment(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewNumber(100),
				),
				NewBinary(
					NewCellAddr(layout.NewPosition(1, 1), false, false),
					NewNumber(2),
					op.Mul,
				),
			}),
		},
	}
	p := NewParser(ScriptGrammar())
	for _, c := range tests {
		expr, err := p.ParseString(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expr: %s", c.Expr, err)
			continue
		}
		got, ok := expr.(Script)
		if !ok {
			t.Errorf("%s: expected script, got %T", c.Expr, expr)
			continue
		}
		assertEqualExpr(t, c.Want, got)
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
	case and:
		g, ok := got.(and)
		if !ok {
			t.Errorf("and expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.left, g.left)
		assertEqualExpr(t, w.right, g.right)
	case or:
		g, ok := got.(or)
		if !ok {
			t.Errorf("or expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.left, g.left)
		assertEqualExpr(t, w.right, g.right)
	case not:
		g, ok := got.(not)
		if !ok {
			t.Errorf("not expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.expr, g.expr)
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
	case postfix:
		g, ok := got.(postfix)
		if !ok {
			t.Errorf("postfix expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.expr, g.expr)
		if g.op != w.op {
			t.Errorf("unary operator mismatched!")
		}
	case deferred:
		g, ok := got.(deferred)
		if !ok {
			t.Errorf("deferred expression expected but got %T", got)
			return
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
	case access:
		g, ok := got.(access)
		if !ok {
			t.Errorf("access expression expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.expr, g.expr)
		if w.prop != g.prop {
			t.Errorf("property mismatched! want %s, got %s", w.prop, g.prop)
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
	case slice:
		g, ok := got.(slice)
		if !ok {
			t.Errorf("slice expected but got %T", got)
			return
		}
		assertEqualExpr(t, w.view, g.view)
		assertEqualExpr(t, w.expr, g.expr)
	case columnsSlice:
		g, ok := got.(columnsSlice)
		if !ok {
			t.Errorf("columns slice expected but got %T", got)
			return
		}
		if len(w.columns) != len(g.columns) {
			t.Errorf("selected columns count mismatched! want %d, got %d", len(w.columns), len(g.columns))
		}
		for i := range w.columns {
			if w.columns[i].from != g.columns[i].from {
				t.Errorf("column \"from\" mismatched! want %d, got %d", w.columns[i].from, g.columns[i].from)
			}
			if w.columns[i].to != g.columns[i].to {
				t.Errorf("column \"to\" mismatched! want %d, got %d", w.columns[i].to, g.columns[i].to)
			}
			if w.columns[i].step != g.columns[i].step {
				t.Errorf("columns \"step\" mismatched! want %d, got %d", w.columns[i].step, g.columns[i].step)
			}
		}
	case Script:
		g, ok := got.(Script)
		if !ok {
			t.Errorf("script expected but got %T", got)
			return
		}
		if len(w.Body) != len(g.Body) {
			t.Errorf("expressions count mismatched! want %d, got %d", len(w.Body), len(g.Body))
			return
		}
		for i := range w.Body {
			assertEqualExpr(t, w.Body[i], g.Body[i])
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

func TestPrecedences(t *testing.T) {
	tests := []struct{
		Expr string
		Want string
	}{
		{
			Expr: "A1 + 2 * A3",
			Want: b(c("A1"), b(n(2), c("A3"), "*"), "+"),
		},
		{
			Expr: "(A1+2) * A3",
			Want: b(b(c("A1"), n(2), "+"), c("A3"), "*"),
		},
		{
			Expr: "1 * 2 * 3",
			Want: b(b(n(1), n(2), "*"), n(3), "*"),
		},
		{
			Expr: "1 + 2 + 3",
			Want: b(b(n(1), n(2), "+"), n(3), "+"),
		},
		{
			Expr: "-1 + 2 + -3",
			Want: b(b(u(n(1), "-"), n(2), "+"), u(n(3), "-"), "+"),
		},
	}
	p := NewParser(ScriptGrammar())
	for _, c := range tests {
		expr, err := p.ParseString(c.Expr)
		if err != nil {
			t.Errorf("%s: fail to parse expr: %s", c.Expr, err)
			continue
		}
		got := DumpExpr(unwrapScriptExpr(expr))
		if c.Want != got {
			t.Errorf("%s: AST dump mismatched! want %s, got %s", c.Expr, c.Want, got)
		}
	}
}

func l(s string) string {
    return fmt.Sprintf("literal(%s)", s)
}

func n(v int) string {
    return fmt.Sprintf("number(%d)", v)
}

func c(name string) string {
    return fmt.Sprintf("cell(%s, false, false)", name)
}

func b(left, right, op string) string {
    return fmt.Sprintf("binary(%s, %s, %s)", left, right, op)
}

func u(left, op string) string {
    return fmt.Sprintf("unary(%s, %s)", left, op)
}
