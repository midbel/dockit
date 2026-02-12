package eval

import (
	"fmt"
	"io"

	"github.com/midbel/dockit/formula/types"
	"github.com/midbel/dockit/value"
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

}
