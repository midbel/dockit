package builtins

import (
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/value"
)

func typeOf(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, ErrArity
	}
	res := args[0].Type()
	return types.Text(res), nil
}