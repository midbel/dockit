package oxml

import "fmt"

var builtins = map[string]func([]Value) (Value, error){
	"sum":     execSum,
	"average": execAvg,
	"count":   execCount,
	"min":     execMin,
	"max":     execMax,
}

func execMin(args []Value) (Value, error) {
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

func execMax(args []Value) (Value, error) {
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

func execSum(args []Value) (Value, error) {
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += float64(args[i].(Float))
	}
	return Float(total), nil
}

func execAvg(args []Value) (Value, error) {
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

func execCount(args []Value) (Value, error) {
	return nil, nil
}
