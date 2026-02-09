package value

import (
	"fmt"

	"github.com/midbel/dockit/layout"
)

type Array struct {
	Data [][]ScalarValue
}

func NewArray(data [][]ScalarValue) ArrayValue {
	return Array{
		Data: data,
	}
}

func (a Array) Type() string {
	dim := a.Dimension()
	return fmt.Sprintf("array(%d, %d)", dim.Lines, dim.Columns)
}

func (Array) Kind() ValueKind {
	return KindArray
}

func (Array) String() string {
	return ""
}

func (a Array) Dimension() layout.Dimension {
	var (
		d layout.Dimension
		n = len(a.Data)
	)
	if n > 0 {
		d.Lines = int64(n)
		d.Columns = int64(len(a.Data[0]))
	}
	return d
}

func (a Array) At(row, col int) ScalarValue {
	if len(a.Data) == 0 || row >= len(a.Data) {
		return nil
	}
	v := a.Data[row]
	if len(v) == 0 || col >= len(v) {
		return nil
	}
	return a.Data[row][col]
}

func (a Array) SetAt(row, col int, val ScalarValue) {
	if len(a.Data) == 0 || row >= len(a.Data) {
		return
	}
	v := a.Data[row]
	if len(v) == 0 || col >= len(v) {
		return
	}
	a.Data[row][col] = val
}

func (a Array) Apply(do func(ScalarValue) (ScalarValue, error)) error {
	if len(a.Data) == 0 {
		return nil
	}
	dim := a.Dimension()
	for i := range dim.Lines {
		for j := range dim.Columns {
			v, err := do(a.At(int(i), int(j)))
			if err != nil {
				return err
			}
			a.SetAt(int(i), int(j), v)
		}
	}
	return nil
}

func (a Array) ApplyArray(other Array, do func(ScalarValue, ScalarValue) (ScalarValue, error)) (Value, error) {
	var (
		dleft  = a.Dimension()
		dright = other.Dimension()
		dim    = dleft.Max(dright)
		data   = make([][]ScalarValue, dim.Lines)
	)
	for i := range data {
		data[i] = make([]ScalarValue, dim.Columns)
	}
	for i := range dim.Lines {
		for j := range dim.Columns {
			var (
				left  = a.At(int(i%dleft.Lines), int(j%dleft.Columns))
				right = other.At(int(i%dright.Lines), int(j%dright.Columns))
			)

			v, err := do(left, right)
			if err != nil {
				return ErrValue, err
			}

			data[i][j] = v
		}
	}
	return NewArray(data), nil
}
