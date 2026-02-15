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

func PrintValue(w io.Writer, rows, cols int) Printer {
	return valuePrinter{
		w:    w,
		rows: rows,
		cols: cols,
	}
}

func DebugValue(w io.Writer, rows, cols int) Printer {
	return debugPrinter{
		w:    w,
		rows: rows,
		cols: cols,
	}
}

type valuePrinter struct {
	w    io.Writer
	cols int
	rows int
}

func (p valuePrinter) Print(v value.Value) {
	switch v := v.(type) {
	case value.ScalarValue:
		p.printScalar(v)
	case value.ArrayValue:
		p.printArray(v)
	case *types.View:
		p.printView(v)
	case *types.InspectValue:
		p.printInspect(v)
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

func (p valuePrinter) printInspect(v *types.InspectValue) {

}

type debugPrinter struct {
	w    io.Writer
	cols int
	rows int
}

func (p debugPrinter) Print(v value.Value) {
	switch v := v.(type) {
	case value.ScalarValue:
		p.printScalar(v)
	case value.ArrayValue:
		p.printArray(v)
	case *types.View:
		p.printView(v)
	case *types.InspectValue:
		p.printInspect(v)
	default:
	}
}

func (p debugPrinter) printScalar(v value.ScalarValue) {
	fmt.Fprintf(p.w, "%s(%s)", v.Type(), v.String())
	fmt.Fprintln(p.w)
}

func (p debugPrinter) printArray(v value.ArrayValue) {
	var (
		dim    = v.Dimension()
		writer = bufio.NewWriter(p.w)
	)
	io.WriteString(writer, "array[rows=")
	io.WriteString(writer, strconv.FormatInt(dim.Lines, 10))
	io.WriteString(writer, ", columns=")
	io.WriteString(writer, strconv.FormatInt(dim.Columns, 10))
	io.WriteString(writer, "] [\n")

	for i := range dim.Lines {
		if i > int64(p.rows) {
			break
		}
		io.WriteString(writer, "  ")
		io.WriteString(writer, "[")
		for j := range dim.Columns {
			if j > 0 {
				io.WriteString(writer, ", ")
			}
			if j > maxCols {
				io.WriteString(writer, "...")
				break
			}
			io.WriteString(writer, v.At(int(i), int(j)).String())
		}
		io.WriteString(writer, "],\n")
	}

	io.WriteString(writer, "]\n")

	writer.Flush()
}

func (p debugPrinter) printView(v *types.View) {
	var (
		view      = v.View()
		bounds    = view.Bounds()
		writer    = bufio.NewWriter(p.w)
		cols      = bounds.Width()
		rows      = bounds.Height()
		truncated = rows > int64(p.rows)
		data      = make([][]string, 0, min(rows, int64(p.rows)))
	)
	if rows == 0 || cols == 0 {
		return
	}

	next, stop := iter.Pull(view.Rows())
	defer stop()

	var (
		first, _ = next()
		size     = min(len(first), p.cols)
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
		if !ok || len(data) >= p.rows {
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
	io.WriteString(writer, "] (")
	for i := range data {
		io.WriteString(writer, "\n  [")
		if i+1 < 10 {
			io.WriteString(writer, "0")
		}
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
		io.WriteString(writer, "\n  ... (")
		io.WriteString(writer, strconv.FormatInt(rows-int64(p.rows), 10))
		io.WriteString(writer, " more rows)\n")
	}
	io.WriteString(writer, ")\n")
	writer.Flush()
}

func (p debugPrinter) printInspect(v *types.InspectValue) {
	var (
		writer = bufio.NewWriter(p.w)
		lino   int
	)
	io.WriteString(writer, "inspect[type=")
	io.WriteString(writer, v.Type())
	io.WriteString(writer, "] [\n")

	for n, v := range v.Values() {
		lino++
		io.WriteString(writer, "  ")
		io.WriteString(writer, "[")
		if lino < 10 {
			io.WriteString(writer, "0")
		}
		io.WriteString(writer, strconv.Itoa(lino))
		io.WriteString(writer, "] ")
		writeValue(writer, n, 12)
		io.WriteString(writer, " => ")
		io.WriteString(writer, v.String())
		io.WriteString(writer, "\n")
	}
	io.WriteString(writer, "]\n")
	writer.Flush()
}

func writeValue(writer io.Writer, str string, size int) {
	io.WriteString(writer, str)
	for range size - len(str) {
		io.WriteString(writer, " ")
	}
}
