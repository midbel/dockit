package types

import (
	"github.com/midbel/dockit/value"
)

type ExprPredicate struct {
	expr value.Formula
}

func NewExprPredicate(expr value.Formula) value.Predicate {
	return ExprPredicate{
		expr: expr,
	}
}

func (p ExprPredicate) Test(ctx value.Context) bool {
	if p.expr == nil {
		return false
	}
	val := p.expr.Eval(ctx)
	return value.True(val)
}
