package builtins

import (
	"github.com/midbel/dockit/value"
)

func Any(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return value.Boolean(false), nil
}

func All(predicate value.Predicate, rg value.Value) (value.Value, error) {
	return value.Boolean(false), nil
}
