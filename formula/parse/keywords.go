package parse

const (
	kwImport  = "import"
	kwUse     = "use"
	kwUsing   = "using"
	kwWith    = "with"
	kwPrint   = "print"
	kwExport  = "export"
	kwDefault = "default"
	kwFrom    = "from"
	kwIn      = "in"
	kwAs      = "as"
	kwTo      = "to"
	kwEnd     = "end"
	kwRo      = "ro"
	kwRw      = "rw"
	kwLock    = "lock"
	kwUnlock  = "unlock"
	kwPush    = "push"
	kwPop     = "pop"
	kwClear   = "clear"
	kwAnd     = "and"
	kwOr      = "or"
	kwNot     = "not"
)

func isKeyword(str string) bool {
	switch str {
	case kwImport:
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
	case kwEnd:
	case kwRo:
	case kwRw:
	case kwLock:
	case kwUnlock:
	case kwPush:
	case kwPop:
	case kwClear:
	case kwAnd:
	case kwOr:
	case kwNot:
	default:
		return false
	}
	return true
}
