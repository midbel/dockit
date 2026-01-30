package op

type Op rune

const (
	Invalid Op = 0

	EOF Op = 1 << iota
	Eol
	Keyword
	Ident
	Number
	Literal
	Comment
	Assign
	Add
	Sub
	Mul
	Div
	Percent
	Pow
	Concat
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	Comma
	Dot
	Begin
	End
	RangeRef
	SheetRef
)

const (
	groupTok Op = 1 << iota
	propTok
)

const (
	BegGrp  = groupTok | Begin
	EndGrp  = groupTok | End
	BegProp = propTok | Begin
	EndProp = propTok | End
)

const (
	AddAssign    = Add | Assign
	DivAssign    = Div | Assign
	SubAssign    = Sub | Assign
	MulAssign    = Mul | Assign
	PowAssign    = Pow | Assign
	ConcatAssign = Concat | Assign
)

var mapping = map[Op]string{
	Add:     "+",
	Sub:     "-",
	Mul:     "*",
	Pow:     "^",
	Div:     "/",
	Percent: "%",
	Concat:  "&",
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
