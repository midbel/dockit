package eval

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strconv"

	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/value"
)

const (
	maxCols = 10
	maxRows = 25
)

type PrintMode int

const (
	PrintDefault PrintMode = 1 << iota
	PrintDebug
)

type Printer interface {
	Print(value.Value)
}

func PrintValue(w io.Writer) Printer {
	return valuePrinter{
		w: w,
	}
}

func DebugValue(w io.Writer) Printer {
	return debugPrinter{
		w: w,
	}
}

type valuePrinter struct {
	w io.Writer
}

func (p valuePrinter) Print(v value.Value) {
	switch v := v.(type) {
	case value.ScalarValue:
		p.printScalar(v)
	case value.ArrayValue:
		p.printArray(v)
	case *types.View:
		p.printView(v)
	default:
	}
}

func (p valuePrinter) printScalar(v value.ScalarValue) {
	fmt.Fprintln(p.w, v.String())
}

func (p valuePrinter) printArray(v value.ArrayValue) {

}

func (p valuePrinter) printView(v *types.View) {

}

type debugPrinter struct {
	w io.Writer
}

func (p debugPrinter) Print(v value.Value) {
	switch v := v.(type) {
	case value.ScalarValue:
		p.printScalar(v)
	case value.ArrayValue:
		p.printArray(v)
	case *types.View:
		p.printView(v)
	default:
	}
}

func (p debugPrinter) printScalar(v value.ScalarValue) {
	fmt.Fprintf(p.w, "%s(%s)", v.Type(), v.String())
	fmt.Fprintln(p.w)
}

func (p debugPrinter) printArray(v value.ArrayValue) {

}

func (p debugPrinter) printView(v *types.View) {
	var (
		view      = v.View()
		bounds    = view.Bounds()
		writer    = bufio.NewWriter(p.w)
		cols      = bounds.Width()
		rows      = bounds.Height()
		truncated = rows > maxRows
		data      = make([][]string, 0, min(rows, maxRows))
	)
	if rows == 0 || cols == 0 {
		return
	}

	next, stop := iter.Pull(view.Rows())
	defer stop()

	var (
		first, _ = next()
		size     = min(len(first), maxCols)
		padding  = make([]int, size)
		row      = make([]string, size)
	)
	for i := range size {
		row[i] = first[i].String()
		padding[i] = max(padding[i], len(row[i]))
	}
	data = append(data, row)

	for {
		r, ok := next()
		if !ok || len(data) >= maxRows {
			break
		}
		row = make([]string, size)
		for i := range size {
			row[i] = r[i].String()
			padding[i] = max(padding[i], len(row[i]))
		}
		data = append(data, row)
	}

	io.WriteString(writer, "view[rows=")
	io.WriteString(writer, strconv.FormatInt(rows, 10))
	io.WriteString(writer, ", columns=")
	io.WriteString(writer, strconv.FormatInt(cols, 10))
	io.WriteString(writer, "]")
	for i := range data {
		io.WriteString(writer, "\n")
		io.WriteString(writer, "[")
		io.WriteString(writer, strconv.Itoa(i+1))
		io.WriteString(writer, "] ")
		for j := range data[i] {
			if j > 0 {
				io.WriteString(writer, " | ")
			}
			writeValue(writer, data[i][j], padding[j])
		}
	}

	if truncated {
		io.WriteString(writer, "\n... (")
		io.WriteString(writer, strconv.FormatInt(rows-maxRows, 10))
		io.WriteString(writer, " more rows)")
	}
	io.WriteString(writer, "\n")
	writer.Flush()
}

func writeValue(writer io.Writer, str string, size int) {
	io.WriteString(writer, str)
	for range size - len(str) {
		io.WriteString(writer, " ")
	}
}
