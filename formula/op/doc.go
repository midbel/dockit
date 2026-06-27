// Package op defines lexical and syntactic operator tokens for Dockit formulas
// and scripts.
//
// Op values are shared by scanners, parsers, formatters, and evaluators. They
// include structural tokens such as EOF and delimiters, literal classes such as
// identifiers and numbers, assignment operators, arithmetic operators,
// comparisons, logical operators, cell/range separators, and script-specific
// markers.
//
// Symbol returns the display form for operators that have a direct textual
// representation. It returns an empty string for token classes that are not
// simple printable operators.
package op
