package grid

import (
	"slices"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

const (
	CallComplexity      = 2
	LiteralComplexity   = 0
	ReferenceComplexity = 1
	RangeComplexity     = 2
	BinaryComplexity    = 1
	UnaryComplexity     = 1
)

type Top[T any] struct {
	Item  T
	Count int
}

type FormulaStats struct {
	layout.Position
	Formula string

	Builtins   map[string]int
	Literals   int
	References int
	Complexity int
	MaxDepth   int

	Deps []layout.Position
}

type ViewStats struct {
	Name string

	Cells     int
	Formulas  int
	Errors    int
	Constants int

	Complexity int
	MaxDepth   int

	Builtins map[string]int
	Refs     map[layout.Position]int
}

func (s ViewStats) TopBuiltins(n int) []Top[string] {
	var list []Top[string]
	for b, c := range s.Builtins {
		t := Top[string]{
			Item:  b,
			Count: c,
		}
		list = append(list, t)
	}
	slices.SortFunc(list, func(t1, t2 Top[string]) int {
		return t2.Count - t1.Count
	})
	if len(list) <= n {
		return list
	}
	return list[:n]
}

func (s ViewStats) TopRefs(n int) []Top[layout.Position] {
	var list []Top[layout.Position]
	for p, c := range s.Refs {
		t := Top[layout.Position]{
			Item:  p,
			Count: c,
		}
		list = append(list, t)
	}
	slices.SortFunc(list, func(t1, t2 Top[layout.Position]) int {
		return t2.Count - t1.Count
	})
	if len(list) <= n {
		return list
	}
	return list[:n]
}

func AnalyzeView(view View) ViewStats {
	stat := ViewStats{
		Name:     view.Name(),
		Builtins: make(map[string]int),
		Refs:     make(map[layout.Position]int),
	}

	b := view.Bounds()
	for p := range b.Positions() {
		var (
			c, _ = view.Cell(p)
			v    = c.Value()
		)

		stat.Cells++
		if value.IsError(v) {
			stat.Errors++
		}
		f, ok := c.(interface{ Formula() value.Formula })
		if ok {
			sub := f.Formula()
			if sub != nil {
				stat.Formulas++
				fs := AnalyzeFormula(sub)
				for n, c := range fs.Builtins {
					stat.Builtins[n] += c
				}
				for _, d := range fs.Deps {
					stat.Refs[d]++
				}
				// stat.Formulas = append(stat.Formulas, fs)
				stat.MaxDepth = max(stat.MaxDepth, fs.MaxDepth)
				stat.Complexity = max(stat.Complexity, fs.Complexity)
			}
		} else {
			stat.Constants++
		}
	}
	return stat
}

func AnalyzeFormula(form value.Formula) FormulaStats {
	return Walk(form)
}

func TopComplexFormulas(view View, n int) []FormulaStats {
	var (
		list []FormulaStats
		rg   = view.Bounds()
	)
	for p := range rg.Positions() {
		c, _ := view.Cell(p)
		f, ok := c.(interface{ Formula() value.Formula })
		if !ok {
			continue
		}
		sub := f.Formula()
		if sub == nil {
			continue
		}
		fs := AnalyzeFormula(sub)
		fs.Position = p
		list = append(list, fs)
	}
	if n > 0 && n < len(list) {
		list = list[:n]
	}
	slices.SortFunc(list, func(s1, s2 FormulaStats) int {
		return s2.Complexity - s1.Complexity
	})
	return list
}

func Walk(f value.Formula) FormulaStats {
	stat := FormulaStats{
		Builtins: make(map[string]int),
	}
	if f == nil {
		return stat
	}
	fx, ok := f.(formula)
	if !ok {
		return stat
	}
	walkFormula(fx.expr, 1, &stat)
	stat.Complexity += stat.MaxDepth
	stat.Formula = fx.expr.String()
	// if c, ok := fx.expr.(parse.Call); ok {
	// 	stat.Formula = c.Name().String()
	// } else {
	// 	stat.Formula = fx.expr.String()
	// }
	return stat
}

func walkFormula(expr parse.Expr, depth int, stat *FormulaStats) {
	if depth > stat.MaxDepth {
		stat.MaxDepth = depth
	}
	var complexity int

	switch e := expr.(type) {
	default:
	case parse.Binary:
		walkFormula(e.Left(), depth+1, stat)
		walkFormula(e.Right(), depth+1, stat)

		complexity = BinaryComplexity
	case parse.Unary:
		walkFormula(e.Expr(), depth+1, stat)

		complexity = UnaryComplexity
	case parse.Literal, parse.Number:
		stat.Literals++

		complexity = LiteralComplexity
	case parse.Call:
		id := e.Name()
		if id, ok := id.(parse.Identifier); ok {
			stat.Builtins[id.Ident()]++
		}
		for _, a := range e.Args() {
			walkFormula(a, depth+1, stat)
		}

		complexity = CallComplexity
	case parse.CellAddr:
		stat.References++
		stat.Deps = append(stat.Deps, e.Position)

		complexity = ReferenceComplexity
	case parse.RangeAddr:
		stat.References += 2
		stat.Deps = append(stat.Deps, e.StartAt().Position)
		stat.Deps = append(stat.Deps, e.EndAt().Position)

		complexity = RangeComplexity
	}
	stat.Complexity += complexity
}
