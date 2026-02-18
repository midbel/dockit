package eval

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"math"
	"slices"
	"strconv"
	"strings"

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

type Formatter interface {
	Format(value.Value) (string, error)
}

type valueFormatter struct {
	list []Formatter
}

func (valueFormatter) Format(v value.Value) (string, error) {
	return "", nil
}

type strFormatter struct{}

func (strFormatter) Format(v value.Value) (string, error) {
	return v.String(), nil
}

type dateFormatter struct{}

func (f dateFormatter) Format(v value.Value) (string, error) {
	return "", nil
}

type numberFormatter struct {
	minInt int
	maxInt int
	minDec int
	maxDec int

	signAlways  bool
	hasGrouping bool
	hasDecimal  bool

	decimalSep  byte
	thousandSep byte
}

func ParseNumberFormatter(pattern string) (Formatter, error) {
	if pattern == "" || pattern == "." || pattern == "-" || pattern == "+" {
		return nil, fmt.Errorf("invalid pattern given")
	}
	var (
		nf     numberFormatter
		left   string
		right  string
		zeroes = true
	)
	nf.decimalSep = '.'
	nf.thousandSep = ','

	left, right, nf.hasDecimal = strings.Cut(pattern, ".")

	if left == "" || left == "-" || left == "+" {
		return nil, fmt.Errorf("invalid pattern given")
	}

	for i := 0; i < len(right); i++ {
		if zeroes && right[i] == '0' {
			nf.minDec++
			nf.maxDec++
		} else if right[i] == '#' {
			zeroes = false
			nf.maxDec++
		} else {
			return nil, fmt.Errorf("unexpected character in fractional part pattern")
		}
	}

	if left[0] == '+' {
		nf.signAlways = true
		left = left[1:]
	}

	zeroes = true
	for i := len(left) - 1; i >= 0; i-- {
		if left[i] == ',' {
			nf.hasGrouping = true
			continue
		}

		if zeroes && left[i] == '0' {
			nf.minInt++
			nf.maxInt++
		} else if left[i] == '#' {
			zeroes = false
			nf.maxInt++
		} else {
			return nil, fmt.Errorf("unexpected character in integral part pattern")
		}
	}

	return nf, nil
}

func (nf numberFormatter) Format(v value.Value) (string, error) {
	vf, ok := v.(value.Float)
	if !ok {
		return "", fmt.Errorf("value is not a number")
	}

	var (
		scale      = math.Pow10(nf.maxDec)
		rounded    = math.Round(float64(vf)*scale) / scale
		integral   []byte
		fractional []byte
		str        = strconv.FormatFloat(rounded, 'f', nf.maxDec, 64)
		signed     = math.Signbit(float64(vf))
	)
	left, right, _ := strings.Cut(str, ".")
	if nf.maxDec > 0 {
		fractional = make([]byte, nf.maxDec)
		for i := 0; i < nf.maxDec; i++ {
			fractional[i] = '0'
		}
		copy(fractional, right)
	}
	for len(fractional) > nf.minDec && fractional[len(fractional)-1] == '0' {
		fractional = fractional[:len(fractional)-1]
	}
	if signed {
		left = left[1:]
	}
	integral = []byte(left)
	if z := len(integral); z < nf.minInt {
		tmp := make([]byte, nf.minInt)
		for i := 0; i < nf.minInt; i++ {
			tmp[i] = '0'
		}
		copy(tmp[nf.minInt-z:], integral)
		integral = tmp
	}
	if nf.hasGrouping {
		slices.Reverse(integral)
		var tmp []byte
		for i := 0; i < len(integral); i += 3 {
			if i+3 >= len(integral) {
				tmp = append(tmp, integral[i:]...)
			} else {
				tmp = append(tmp, integral[i:i+3]...)
				tmp = append(tmp, nf.thousandSep)
			}
		}
		slices.Reverse(tmp)
		integral = tmp
	}
	if signed || nf.signAlways {
		tmp := []byte{'+'}
		if signed {
			tmp[0] = '-'
		}
		integral = append(tmp, integral...)
	}
	var all []byte
	if len(fractional) > 0 {
		all = append(integral, nf.decimalSep)
		all = append(all, fractional...)
	} else {
		all = integral
	}
	return string(all), nil
}

type boolFormatter struct{}

func (f boolFormatter) Format(v value.Value) (string, error) {
	return "", nil
}

type Printer interface {
	Print(value.Value)
}

func PrintValue(w io.Writer, rows, cols int) Printer {
	return valuePrinter{
		w:      w,
		rows:   rows,
		cols:   cols,
		format: strFormatter{},
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
	w      io.Writer
	cols   int
	rows   int
	format Formatter
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
	str, err := p.format.Format(v)
	if err != nil {
		str = value.ErrNA.String()
	}
	fmt.Fprintln(p.w, str)
}

func (p valuePrinter) printArray(v value.ArrayValue) {
	writer := bufio.NewWriter(p.w)
	writeArray(writer, v, int64(p.rows), int64(p.cols))
	writer.Flush()
}

func (p valuePrinter) printView(v *types.View) {
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
		for i := 0; i < min(size, len(row)); i++ {
			row[i] = r[i].String()
			padding[i] = max(padding[i], len(row[i]))
		}
		data = append(data, row)
	}

	for i := range data {
		for j := range data[i] {
			if j > 0 {
				io.WriteString(writer, " | ")
			}
			writeValue(writer, data[i][j], padding[j])
		}
		io.WriteString(writer, "\n")
	}
	writeView(writer, data, padding, false)

	if truncated {
		writeTruncate(writer, rows-int64(p.rows))
		io.WriteString(writer, "\n")
	}
	writer.Flush()
}

func (p valuePrinter) printInspect(v *types.InspectValue) {
	var (
		prefix = v.Type()
		props  = make([]string, 0, 5)
		values = make([]string, 0, 5)
	)
	switch prefix {
	case types.InspectKindCell:
		props = append(props, "position", "value", "type")
	case types.InspectKindFile:
		props = append(props, "sheets")
	case types.InspectKindSlice:
		props = append(props, "owner", "type", "rows", "cols")
	case types.InspectKindRange:
		props = append(props, "owner", "rows", "cols")
	case types.InspectKindView:
		props = append(props, "name", "rows", "cols")
	case types.InspectKindPrimitive:
		props = append(props, "type", "value")
	default:
	}
	for _, p := range props {
		x, err := v.Get(p)
		if err == nil {
			str := fmt.Sprintf("%s=%s", p, x.String())
			values = append(values, str)
		}
	}
	fmt.Fprintf(p.w, "%s(%s)\n", prefix, strings.Join(values, ", "))
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
	io.WriteString(writer, "]")

	writeArray(writer, v, int64(p.rows), int64(p.cols))
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
		for i := 0; i < min(size, len(row)); i++ {
			row[i] = r[i].String()
			padding[i] = max(padding[i], len(row[i]))
		}
		data = append(data, row)
	}

	io.WriteString(writer, "view[rows=")
	io.WriteString(writer, strconv.FormatInt(rows, 10))
	io.WriteString(writer, ", columns=")
	io.WriteString(writer, strconv.FormatInt(cols, 10))
	io.WriteString(writer, "] (\n")
	writeView(writer, data, padding, true)

	if truncated {
		io.WriteString(writer, "  ")
		writeTruncate(writer, rows-int64(p.rows))
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

func writeArray(writer io.Writer, arr value.ArrayValue, maxRows, maxCols int64) {
	dim := arr.Dimension()
	io.WriteString(writer, "[\n")
	for i := range dim.Lines {
		if i > maxRows {
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
			io.WriteString(writer, arr.At(int(i), int(j)).String())
		}
		io.WriteString(writer, "],\n")
	}
	io.WriteString(writer, "]\n")
}

func writeTruncate(writer io.Writer, count int64) {
	io.WriteString(writer, "... (")
	io.WriteString(writer, strconv.FormatInt(count, 10))
	io.WriteString(writer, " more rows)\n")
}

func writeView(writer io.Writer, data [][]string, padding []int, index bool) {
	for i := range data {
		if index {
			io.WriteString(writer, "  [")
			if i+1 < 10 {
				io.WriteString(writer, "0")
			}
			io.WriteString(writer, strconv.Itoa(i+1))
			io.WriteString(writer, "] ")
		}
		for j := range data[i] {
			if j > 0 {
				io.WriteString(writer, " | ")
			}
			writeValue(writer, data[i][j], padding[j])
		}
		io.WriteString(writer, "\n")
	}
}
