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
)

func isKeyword(str string) bool {
	switch str {
	case kwAssert:
	case kwElse:
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
	case kwAnd:
	case kwOr:
	case kwNot:
	// case kwMacro:
	// case kwInclude:
	default:
		return false
	}
	return true
}
