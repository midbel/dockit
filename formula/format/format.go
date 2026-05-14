package format

import (
	"io"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
)

type DialectFormat interface {
	Prefix() string
	ArgSeparator() string
	FormatCell(parse.CellAddr) (string, error)
	FormatRange(parse.RangeAddr) (string, error)
}

type odsFormatter struct{}

func (odsFormatter) Prefix() string {
	return "of:"
}

func (odsFormatter) ArgSeparator() string {
	return ";"
}

func (odsFormatter) FormatCell(expr parse.CellAddr) (string, error) {
	var str strings.Builder
	str.WriteString("[")
	str.WriteString(".")
	writeCellAddr(&str, expr)
	str.WriteString("]")
	return str.String(), nil
}

func (odsFormatter) FormatRange(expr parse.RangeAddr) (string, error) {
	var str strings.Builder
	str.WriteString("[")
	str.WriteString(".")
	writeCellAddr(&str, expr.StartAt())
	io.WriteString(&str, ":")
	str.WriteString(".")
	writeCellAddr(&str, expr.EndAt())
	str.WriteString("]")
	return str.String(), nil
}

type oxmlFormatter struct{}

func (oxmlFormatter) Prefix() string {
	return ""
}

func (oxmlFormatter) ArgSeparator() string {
	return ","
}

func (oxmlFormatter) FormatCell(expr parse.CellAddr) (string, error) {
	var str strings.Builder
	writeCellAddr(&str, expr)
	return str.String(), nil
}

func (f oxmlFormatter) FormatRange(expr parse.RangeAddr) (string, error) {
	var str strings.Builder
	writeCellAddr(&str, expr.StartAt())
	io.WriteString(&str, ":")
	writeCellAddr(&str, expr.EndAt())
	return str.String(), nil
}

var (
	Oxml DialectFormat = oxmlFormatter{}
	Ods  DialectFormat = odsFormatter{}
)

func FormatOxml(expr parse.Expr) (string, error) {
	return Format(expr, Oxml)
}

func FormatOds(expr parse.Expr) (string, error) {
	return Format(expr, Ods)
}

func Format(expr parse.Expr, dialect DialectFormat) (string, error) {
	var str strings.Builder
	if pfx := dialect.Prefix(); pfx != "" {
		str.WriteString(pfx)
	}
	str.WriteString("=")
	if err := formatExpr(&str, expr, dialect); err != nil {
		return "", err
	}
	return str.String(), nil
}

func formatExpr(w io.Writer, expr parse.Expr, dialect DialectFormat) error {
	switch expr := expr.(type) {
	default:
	case parse.CellAddr:
		str, err := dialect.FormatCell(expr)
		if err != nil {
			return err
		}
		io.WriteString(w, str)
	case parse.RangeAddr:
		str, err := dialect.FormatRange(expr)
		if err != nil {
			return err
		}
		io.WriteString(w, str)
	case parse.Identifier:
		io.WriteString(w, expr.Ident())
	case parse.Literal:
		io.WriteString(w, "\"")
		io.WriteString(w, expr.Text())
		io.WriteString(w, "\"")
	case parse.Number:
		io.WriteString(w, expr.String())
	case parse.Call:
		if err := formatExpr(w, expr.Name(), dialect); err != nil {
			return err
		}
		io.WriteString(w, "(")
		for i, a := range expr.Args() {
			if i > 0 {
				io.WriteString(w, dialect.ArgSeparator())
				io.WriteString(w, " ")
			}
			if err := formatExpr(w, a, dialect); err != nil {
				return err
			}
		}
		io.WriteString(w, ")")
	case parse.Binary:
		if err := formatExpr(w, expr.Left(), dialect); err != nil {
			return err
		}
		io.WriteString(w, " ")
		io.WriteString(w, op.Symbol(expr.Op()))
		io.WriteString(w, " ")
		if err := formatExpr(w, expr.Right(), dialect); err != nil {
			return err
		}
	case parse.Unary:
		io.WriteString(w, op.Symbol(expr.Op()))
		if err := formatExpr(w, expr.Expr(), dialect); err != nil {
			return err
		}
	case parse.Postfix:
		if err := formatExpr(w, expr.Expr(), dialect); err != nil {
			return err
		}
		io.WriteString(w, op.Symbol(expr.Op()))
	case parse.CellAccess:
	}
	return nil
}

func writeCellAddr(w io.Writer, expr parse.CellAddr) {
	if expr.AbsCol {
		io.WriteString(w, "$")
	}
	column := expr.Column
	for column > 0 {
		column--
		io.WriteString(w, string(rune('A')+rune(column%26)))
		column /= 26
	}
	if expr.AbsRow {
		io.WriteString(w, "$")
	}
	io.WriteString(w, strconv.FormatInt(expr.Line, 10))
}
