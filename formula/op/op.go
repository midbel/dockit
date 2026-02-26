package op

type Op rune

const (
	Invalid Op = 0

	EOF Op = -(1 + iota)
	Eol
	Directive
	Pragma
	Keyword
	Ident
	Cell
	Number
	Literal
	Comment
	Assign
	AddAssign
	DivAssign
	SubAssign
	MulAssign
	PowAssign
	ConcatAssign
	Add
	Sub
	Mul
	Div
	Percent
	Pow
	Concat
	Union
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	And
	Or
	Not
	Comma
	Semi
	Dot
	BegGrp
	EndGrp
	BegProp
	EndProp
	RangeRef
	SheetRef
)

var mapping = map[Op]string{
	Add:     "+",
	Sub:     "-",
	Mul:     "*",
	Pow:     "^",
	Div:     "/",
	Percent: "%",
	Concat:  "&",
	Union:   "|",
	Eq:      "=",
	Ne:      "<>",
	Lt:      "<",
	Le:      "<=",
	Gt:      ">",
	Ge:      ">=",
}

func Symbol(oper Op) string {
	return mapping[oper]
}
