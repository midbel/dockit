package eval

import (
	"fmt"

	"github.com/midbel/dockit/formula/op"
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Literal string
	Type    op.Op
	Position
}

func (t Token) String() string {
	var str string
	switch t.Type {
	case op.Invalid:
		return "<invalid>"
	case op.EOF:
		return "<eof>"
	case op.Eol:
		return "<eol>"
	case op.Keyword:
		str = "keyword"
	case op.Ident:
		str = "identifier"
	case op.Number:
		str = "number"
	case op.Literal:
		str = "literal"
	case op.Comment:
		str = "comment"
	case op.Assign:
		return "<assignment>"
	case op.Add:
		return "<add>"
	case op.Sub:
		return "<subtract>"
	case op.Mul:
		return "<multiply>"
	case op.Div:
		return "<divide>"
	case op.Percent:
		return "<percent>"
	case op.Pow:
		return "<power>"
	case op.Concat:
		return "<concat>"
	case op.Eq:
		return "<equal>"
	case op.Ne:
		return "<notequal>"
	case op.Lt:
		return "<lesser>"
	case op.Le:
		return "<lesseq>"
	case op.Gt:
		return "<greater>"
	case op.Ge:
		return "<greateq>"
	case op.Comma:
		return "<comma>"
	case op.Dot:
		return "<dot>"
	case op.BegGrp:
		return "<beg-group>"
	case op.EndGrp:
		return "<end-group>"
	case op.BegProp:
		return "<beg-prop>"
	case op.EndProp:
		return "<end-prop>"
	case op.BegBlock:
		return "<beg-block>"
	case op.EndBlock:
		return "<end-block>"
	case op.RangeRef:
		return "<range>"
	case op.SheetRef:
		return "<sheet>"
	case op.AddAssign:
		return "<add-assign>"
	case op.DivAssign:
		return "<div-assign>"
	case op.SubAssign:
		return "<sub-assign>"
	case op.MulAssign:
		return "<mul-assign>"
	case op.PowAssign:
		return "<pow-assign>"
	case op.ConcatAssign:
		return "<concat-assign>"
	}
	return fmt.Sprintf("%s(%s)", str, t.Literal)
}
