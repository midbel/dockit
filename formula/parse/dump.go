package parse

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/midbel/dockit/formula/op"
)

func DumpExpr(expr Expr) string {
	var buf bytes.Buffer
	dumpExpr(&buf, expr)
	return buf.String()
}

func dumpExpr(w io.Writer, expr Expr) {
	switch e := expr.(type) {
	case Identifier:
		io.WriteString(w, "identifier(")
		io.WriteString(w, e.name)
		io.WriteString(w, ")")
	case Literal:
		io.WriteString(w, "literal(")
		io.WriteString(w, e.value)
		io.WriteString(w, ")")
	case Number:
		io.WriteString(w, "number(")
		io.WriteString(w, strconv.FormatFloat(e.value, 'f', -1, 64))
		io.WriteString(w, ")")
	case Template:
		io.WriteString(w, "template(")
		for i := range e.expr {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			dumpExpr(w, e.expr[i])
		}
		io.WriteString(w, ")")
	case Binary:
		io.WriteString(w, "binary(")
		dumpExpr(w, e.left)
		io.WriteString(w, ", ")
		dumpExpr(w, e.right)
		io.WriteString(w, ", ")
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case Unary:
		io.WriteString(w, "unary(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case Spread:
		io.WriteString(w, "spread(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case Not:
		io.WriteString(w, "not(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case And:
		io.WriteString(w, "and(")
		dumpExpr(w, e.left)
		io.WriteString(w, ", ")
		dumpExpr(w, e.right)
		io.WriteString(w, ")")
	case Or:
		io.WriteString(w, "or(")
		dumpExpr(w, e.left)
		io.WriteString(w, ", ")
		dumpExpr(w, e.right)
		io.WriteString(w, ")")
	case Postfix:
		io.WriteString(w, "postfix(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		io.WriteString(w, op.Symbol(e.op))
		io.WriteString(w, ")")
	case Access:
		io.WriteString(w, "access(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ", ")
		io.WriteString(w, e.prop)
		io.WriteString(w, ")")
	case Deferred:
		io.WriteString(w, "deferred(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case Call:
		io.WriteString(w, "call(")
		dumpExpr(w, e.ident)
		io.WriteString(w, ", args: ")
		for i := range e.args {
			if i > 0 {
				io.WriteString(w, ", ")
			}
			dumpExpr(w, e.args[i])
		}
		io.WriteString(w, ")")
	case Assignment:
		io.WriteString(w, "assignment(")
		dumpExpr(w, e.ident)
		io.WriteString(w, ", ")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case CellAddr:
		io.WriteString(w, "cell(")
		io.WriteString(w, e.Position.String())
		io.WriteString(w, ", ")
		io.WriteString(w, strconv.FormatBool(e.AbsCol))
		io.WriteString(w, ", ")
		io.WriteString(w, strconv.FormatBool(e.AbsRow))
		io.WriteString(w, ")")
	case RangeAddr:
		io.WriteString(w, "range(")
		dumpExpr(w, e.startAddr)
		io.WriteString(w, ", ")
		dumpExpr(w, e.endAddr)
		io.WriteString(w, ")")
	case QualifiedCellAddr:
		io.WriteString(w, "qualified(")
		dumpExpr(w, e.path)
		io.WriteString(w, ", ")
		dumpExpr(w, e.addr)
		io.WriteString(w, ")")
	case Slice:
		io.WriteString(w, "slice(")
		dumpExpr(w, e.view)
		io.WriteString(w, ", ")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case ColumnsSlice:
		io.WriteString(w, "selection(")
		for i := range e.columns {
			if i > 0 {
				io.WriteString(w, ",")
			}
			var fix, tix string
			if e.columns[i].from != 0 {
				fix = strconv.Itoa(e.columns[i].from)
			}
			if e.columns[i].to != 0 {
				tix = strconv.Itoa(e.columns[i].to)
			}
			io.WriteString(w, fix)
			if fix != tix {
				io.WriteString(w, ":")
				io.WriteString(w, tix)
			}
		}
		io.WriteString(w, ")")
	case ImportFile:
		io.WriteString(w, "import(")
		io.WriteString(w, e.file)
		if e.format != "" {
			io.WriteString(w, ", format: ")
			io.WriteString(w, e.format)
		}
		if e.alias != "" {
			io.WriteString(w, ", alias: ")
			io.WriteString(w, e.alias)
		}
		if e.defaultFile {
			io.WriteString(w, ", default")
		}
		if e.readOnly {
			io.WriteString(w, ", readonly")
		}
		io.WriteString(w, ")")
	case UseRef:
		io.WriteString(w, "use(")
		io.WriteString(w, e.ident)
		io.WriteString(w, ")")
	case PrintRef:
		io.WriteString(w, "print(")
		dumpExpr(w, e.expr)
		io.WriteString(w, ")")
	case ExportRef:
	case SaveRef:
		io.WriteString(w, "save(")
		io.WriteString(w, ")")
	case Push:
		io.WriteString(w, "push()")
	case Pop:
		io.WriteString(w, "pop()")
	case Clear:
		io.WriteString(w, "clear(")
		io.WriteString(w, e.name)
		io.WriteString(w, ")")
	default:
		io.WriteString(w, fmt.Sprintf("unknown(%T)", e))
	}
}
