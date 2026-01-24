package builtins

import (
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/value"
)

func callAny(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return types.Boolean(false), nil
}

func callAll(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return types.Boolean(false), nil
}

func callCount(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return types.Float(0), nil
}
