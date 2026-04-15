package parse

type Visitor interface {
	VisitScript(Script) error

	VisitIncludeFile(IncludeFile) error
	VisitImportFile(ImportFile) error
	VisitExportRef(ExportRef) error
	VisitPrintRef(PrintRef) error
	VisitUseRef(UseRef) error

	VisitIdentifier(Identifier) error
	VisitLiteral(Literal) error
	VisitNumber(Number) error
	VisitCellAddr(CellAddr) error
	VisitRangeAddr(RangeAddr) error
	VisitTemplate(Template) error
	VisitAccess(Access) error
	VisitSpecial(SpecialAccess) error
	VisitCellAccess(CellAccess) error
	VisitDeferred(Deferred) error
	VisitCall(Call) error
	VisitSlice(Slice) error

	VisitAssert(Assert) error
	VisitBinary(Binary) error
	VisitAssignment(Assignment) error
	VisitPostfix(Postfix) error
	VisitNot(Not) error
	VisitAnd(And) error
	VisitOr(Or) error
	VisitUnary(Unary) error
}

type VisitableExpr interface {
	Accept(Visitor) error
}
