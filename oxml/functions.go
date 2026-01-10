package oxml

import (
	"github.com/midbel/dockit/formula"
	"github.com/midbel/dockit/value"
)

var builtins = map[string]formula.BuiltinFunc{
	"sum":     execSum,
	"average": execAvg,
	"count":   execCount,
	"min":     execMin,
	"max":     execMax,
}

func execMin(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !formula.IsNumber(args[i]) {
			return formula.ErrValue, nil
		}
		v := args[i].(formula.Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = min(res, float64(v))
	}
	return formula.Float(res), nil
}

func execMax(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !formula.IsNumber(args[i]) {
			return formula.ErrValue, nil
		}
		v := args[i].(formula.Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = max(res, float64(v))
	}
	return formula.Float(res), nil
}

func execSum(args []value.Value) (value.Value, error) {
	var total float64
	for i := range args {
		if !formula.IsNumber(args[i]) {
			return formula.ErrValue, nil
		}
		total += float64(args[i].(formula.Float))
	}
	return formula.Float(total), nil
}

func execAvg(args []value.Value) (value.Value, error) {
	if len(args) == 0 {
		return formula.Float(0), nil
	}
	var total float64
	for i := range args {
		if !formula.IsNumber(args[i]) {
			return formula.ErrValue, nil
		}
		total += float64(args[i].(formula.Float))
	}
	return formula.Float(total / float64(len(args))), nil
}

func execCount(args []value.Value) (value.Value, error) {
	return nil, nil
}
