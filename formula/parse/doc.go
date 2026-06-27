// Package parse scans and parses Dockit formulas and scripts.
//
// The package supports multiple input modes. Formula scanners parse OXML-style
// and ODS/OpenFormula-style expressions, while the script scanner parses the
// Dockit scripting language used by the CLI. ParseOxmlFormula and
// ParseOdsFormula are convenient entry points for spreadsheet formulas.
// ScanScript and NewParser are used by higher-level packages to parse complete
// scripts.
//
// Parsed input is represented as small AST node types that implement Expr and,
// for executable scripts, participate in the Visitor interface. Nodes cover
// literals, identifiers, cells, ranges, calls, unary and binary expressions,
// assignments, imports, assertions, printing, export, sheet/view operations,
// slices, and special property access.
//
// The parser is built around a grammar stack and Pratt-style prefix, infix, and
// postfix parse functions. It also extracts script configuration entries before
// evaluation. Some grammar productions are ahead of evaluator support, so a
// syntactically valid script is not always fully executable yet.
package parse
