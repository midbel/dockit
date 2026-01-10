package oxml

import (
	"fmt"

	"github.com/midbel/dockit/value"
)

var builtins = map[string]func([]value.Value) (value.Value, error){
	"sum":     execSum,
	"average": execAvg,
	"count":   execCount,
	"min":     execMin,
	"max":     execMax,
}

func execMin(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		v := args[i].(Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = min(res, float64(v))
	}
	return Float(res), nil
}

func execMax(args []value.Value) (value.Value, error) {
	var res float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		v := args[i].(Float)
		if i == 0 {
			res = float64(v)
			continue
		}
		res = max(res, float64(v))
	}
	return Float(res), nil
}

func execSum(args []value.Value) (value.Value, error) {
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += float64(args[i].(Float))
	}
	return Float(total), nil
}

func execAvg(args []value.Value) (value.Value, error) {
	if len(args) == 0 {
		return Float(0), nil
	}
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += float64(args[i].(Float))
	}
	return Float(total / float64(len(args))), nil
}

func execCount(args []value.Value) (value.Value, error) {
	return nil, nil
}
