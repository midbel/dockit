package parse

type Visitor interface {
	VisitLockRef(LockRef) error
	VisitUnlockRef(UnlockRef) error
	VisitUseRef(UseRef) error
	VisitImportFile(ImportFile) error
	VisitExportRef(ExportRef) error
	VisitPrintRef(PrintRef) error
	VisitAccess(Access) error
	VisitTemplate(Template) error
	VisitDeferred(Deferred) error
	VisitAssignment(Assignment) error
	VisitBinary(Binary) error
	VisitPostfix(Postfix) error
	VisitNot(Not) error
	VisitAnd(And) error
	VisitOr(Or) error
	VisitUnary(Unary) error
	VisitLiteral(Literal) error
	VisitNumber(Number) error
	VisitCall(Call) error
	VisitClear(Clear) error
	VisitSlice(Slice) error
	VisitIdentifier(Identifier) error
	VisitQualifiedCellAddr(QualifiedCellAddr) error
	VisitCellAddr(CellAddr) error
	VisitRangeAddr(RangeAddr) error
}

type VisitableExpr interface {
	Accept(Visitor) error
}
