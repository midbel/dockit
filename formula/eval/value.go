package eval

import (
	"github.com/midbel/dockit/value"
)

func ScalarToArray(val value.Value, row, col int) value.Value {
	scalar, ok := val.(value.ScalarValue)
	if !ok {
		return value.NewArray(nil)
	}
	arr := make([][]value.ScalarValue, row)
	for i := range row {
		arr[i] = make([]value.ScalarValue, col)
		for j := range col {
			arr[i][j] = scalar
		}
	}
	return value.NewArray(arr)
} 