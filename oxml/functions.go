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
		if i == 0 {
			res = args[i].(float64)
			continue
		}
		res = min(res, args[i].(float64))
	}
	return res, nil
}

func execMax(args []Value) (Value, error) {
	var res float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		if i == 0 {
			res = args[i].(float64)
			continue
		}
		res = max(res, args[i].(float64))
	}
	return res, nil
}

func execSum(args []Value) (Value, error) {
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += args[i].(float64)
	}
	return total, nil
}

func execAvg(args []Value) (Value, error) {
	if len(args) == 0 {
		return 0, nil
	}
	var total float64
	for i := range args {
		if !isNumber(args[i]) {
			return nil, fmt.Errorf("number expected")
		}
		total += args[i].(float64)
	}
	return total / float64(len(args)), nil
}

func execCount(args []Value) (Value, error) {
	return nil, nil
}
