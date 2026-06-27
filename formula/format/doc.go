// Package format renders parsed formula expressions back into spreadsheet
// formula syntax.
//
// Format walks formula/parse AST nodes and writes a textual formula using a
// DialectFormat. The package currently provides Oxml and Ods dialects, exposed
// through FormatOxml and FormatOds helpers. The dialect controls formula
// prefixes, argument separators, and address/range formatting.
//
// The formatter covers the core expression nodes used by formula parsing:
// identifiers, literals, numbers, calls, unary and binary operators, postfix
// operators, cells, and ranges. It is intended for formula expressions, not for
// full Dockit scripts.
package format
