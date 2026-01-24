package builtins

import (
	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/value"
)

func execMin(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !types.IsNumber(args[i]) {
			return types.ErrValue, nil
		}
		v := args[i].(types.Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = min(res, float64(v))
	}
	return types.Float(res), nil
}

func execMax(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !types.IsNumber(args[i]) {
			return types.ErrValue, nil
		}
		v := args[i].(types.Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = max(res, float64(v))
	}
	return types.Float(res), nil
}

func execSum(args []value.Value) (value.Value, error) {
	var total float64
	for i := range args {
		if !types.IsNumber(args[i]) {
			return types.ErrValue, nil
		}
		total += float64(args[i].(types.Float))
	}
	return types.Float(total), nil
}

func execAvg(args []value.Value) (value.Value, error) {
	if len(args) == 0 {
		return types.Float(0), nil
	}
	var total float64
	for i := range args {
		if !types.IsNumber(args[i]) {
			return types.ErrValue, nil
		}
		total += float64(args[i].(types.Float))
	}
	return types.Float(total / float64(len(args))), nil
}

func execCount(args []value.Value) (value.Value, error) {
	return nil, nil
}
