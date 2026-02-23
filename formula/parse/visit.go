package parse

type Visitor[T any] interface {
	VisitPush(Push) (T, error)
	VisitPop(Pop) (T, error)
	VisitLockRef(LockRef) (T, error)
	VisitUnlockRef(UnlockRef) (T, error)
	VisitUseRef(UseRef) (T, error)
	VisitImportFile(ImportFile) (T, error)
	VisitPrintRef(PrintRef) (T, error)
	VisitExportRef(ExportRef) (T, error)
	VisitMacroDef(MacroDef) (T, error)
	VisitAccess(Access) (T, error)
	VisitTemplate(Template) (T, error)
	VisitDeferred(Deferred) (T, error)
	VisitAssignment(Assignment) (T, error)
	VisitBinary(Binary) (T, error)
	VisitPostfix(Postfix) (T, error)
	VisitNot(Not) (T, error)
	VisitAnd(And) (T, error)
	VisitOr(Or) (T, error)
	VisitSpread(Spread) (T, error)
	VisitUnary(Unary) (T, error)
	VisitLiteral(Literal) (T, error)
	VisitNumber(Number) (T, error)
	VisitCall(Call) (T, error)
	VisitClear(Clear) (T, error)
	VisitSlice(Slice) (T, error)
	VisitIdentifier(Identifier) (T, error)
	VisitQualifiedCellAddr(QualifiedCellAddr) (T, error)
	VisitCellAddr(CellAddr) (T, error)
	VisitRangeAddr(RangeAddr) (T, error)
}

type VisitableExpr[T any] interface {
	Accept(Visitor[T]) (T, error)
}
