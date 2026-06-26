package parse

const (
	kwAlias   = "alias"
	kwImport  = "import"
	kwExport  = "export"
	kwRename  = "rename"
	kwLock    = "lock"
	kwUnlock  = "unlock"
	kwRow     = "row"
	kwRows    = "rows"
	kwColumn  = "column"
	kwColumns = "columns"
	kwFirst = "first"
	kwLast = "last"
	kwSheet   = "sheet"
	kwInsert  = "insert"
	kwRemove  = "remove"
	kwResize  = "resize"
	kwUse     = "use"
	kwUsing   = "using"
	kwWith    = "with"
	kwPrint   = "print"
	kwDefault = "default"
	kwFrom    = "from"
	kwIn      = "in"
	kwAs      = "as"
	kwTo      = "to"
	kwInto    = "into"
	kwRo      = "ro"
	kwRw      = "rw"
	kwAnd     = "and"
	kwOr      = "or"
	kwNot     = "not"
	kwAssert  = "assert"
	kwElse    = "else"
	kwInclude = "include"
	kwMacro   = "macro"
	kwEnd     = "end"
	kwBefore  = "before"
	kwAfter   = "after"
)

func isReserved(str string) bool {
	switch str {
	case kwBefore:
	case kwAfter:
	case kwIn:
	case kwAs:
	case kwTo:
	case kwInto:
	case kwWith:
	case kwFrom:
	default:
		return false
	}
	return true
}

func isKeyword(str string) bool {
	switch str {
	case kwAlias:
	case kwAssert:
	case kwImport:
	case kwRename:
	case kwLock:
	case kwUnlock:
	case kwRow:
	case kwRows:
	case kwColumn:
	case kwColumns:
	case kwFirst:
	case kwLast:
	case kwSheet:
	case kwInsert:
	case kwRemove:
	case kwElse:
	case kwUse:
	case kwUsing:
	case kwWith:
	case kwFrom:
	case kwPrint:
	case kwExport:
	case kwDefault:
	case kwIn:
	case kwAs:
	case kwTo:
	case kwInto:
	case kwEnd:
	case kwRo:
	case kwRw:
	case kwAnd:
	case kwOr:
	case kwNot:
	case kwBefore:
	case kwAfter:
	// case kwMacro:
	// case kwInclude:
	default:
		return false
	}
	return true
}
