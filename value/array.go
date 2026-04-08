package value

import (
	"fmt"
	"iter"
	"slices"

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
	return TypeArray
}

func (a Array) GetData() [][]ScalarValue {
	return a.Data
}

func (a Array) Count() int64 {
	d := a.Dimension()
	return d.Lines * d.Columns
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
		return Empty()
	}
	v := a.Data[row]
	if len(v) == 0 || col >= len(v) {
		return Empty()
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

func (a Array) Values() iter.Seq[ScalarValue] {
	it := func(yield func(ScalarValue) bool) {
		dim := a.Dimension()
		for row := range dim.Lines {
			for col := range dim.Columns {
				ok := yield(a.At(int(row), int(col)))
				if !ok {
					return
				}
			}
		}
	}
	return it
}

func (a Array) Clone() Array {
	other := make([][]ScalarValue, 0, len(a.Data))
	for i := range a.Data {
		other = append(other, slices.Clone(a.Data[i]))
	}
	return NewArray(other).(Array)
}

func (a Array) Apply(do func(ScalarValue) ScalarValue) {
	if len(a.Data) == 0 {
		return
	}
	dim := a.Dimension()
	for i := range dim.Lines {
		for j := range dim.Columns {
			v := do(a.At(int(i), int(j)))
			a.SetAt(int(i), int(j), v)
		}
	}
	return
}

func (a Array) ApplyArray(other Array, do func(ScalarValue, ScalarValue) ScalarValue) Value {
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
			data[i][j] = do(left, right)
		}
	}
	return NewArray(data)
}
