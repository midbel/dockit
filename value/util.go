package value

import (
	"fmt"
)

func ExtractDataFromValue(data Value) ([][]ScalarValue, error) {
	if IsScalar(data) {
		data = ScalarToArray(data, 1, 1)
	}
	return ExtractDataFromArray(data)
}

func ExtractDataFromArray(data Value) ([][]ScalarValue, error) {
	if !IsArray(data) {
		return nil, fmt.Errorf("array expected")
	}
	arr, ok := data.(Array)
	if ok {
		return arr.GetData(), nil
	}
	if arr, ok := data.(interface{ AsArray() ArrayValue }); ok {
		return ExtractDataFromArray(arr.AsArray())
	}
	return nil, fmt.Errorf("array expected")
}

func ScalarToArray(val Value, row, col int) Value {
	scalar, ok := val.(ScalarValue)
	if !ok {
		return NewArray(nil)
	}
	arr := make([][]ScalarValue, row)
	for i := range row {
		arr[i] = make([]ScalarValue, col)
		for j := range col {
			arr[i][j] = scalar
		}
	}
	return NewArray(arr)
}

func ApplyScalarInArray(val ScalarValue, arr ArrayValue, do func(ScalarValue, ScalarValue) (ScalarValue, error)) (Value, error) {
	out := prepareArray(arr)
	for i := range out {
		for j := range out[i] {
			v, err := do(val, arr.At(i, j))
			if err != nil {
				return nil, err
			}
			out[i][j] = v
		}
	}
	return NewArray(out), nil
}

func ApplyArrayWithScalar(arr ArrayValue, val ScalarValue, do func(ScalarValue, ScalarValue) (ScalarValue, error)) (Value, error) {
	out := prepareArray(arr)
	for i := range out {
		for j := range out[i] {
			v, err := do(arr.At(i, j), val)
			if err != nil {
				return nil, err
			}
			out[i][j] = v
		}
	}
	return NewArray(out), nil
}

func prepareArray(arr ArrayValue) [][]ScalarValue {
	var (
		dim = arr.Dimension()
		out = make([][]ScalarValue, dim.Lines)
	)
	for i := range out {
		out[i] = make([]ScalarValue, dim.Columns)
	}
	return out
}
