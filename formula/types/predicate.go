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

func (p ExprPredicate) Test(ctx value.Context) (bool, error) {
	if p.expr == nil {
		return false, nil
	}
	val := p.expr.Eval(ctx)
	if value.IsError(val) {
		return false, nil
	}
	return value.True(val), nil
}
