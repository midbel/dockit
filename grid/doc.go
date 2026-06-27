// Package grid defines Dockit's workbook, sheet, view, cell, and formula
// evaluation primitives.
//
// A File is a workbook-like collection of Views. A View is a rectangular,
// row-oriented dataset that can expose cells by position, stream rows, report
// bounds, and synchronize formula-backed cells against a value.Context.
// MutableView extends View with cell, range, row, and column mutation methods.
//
// The package contains composable view adapters such as read-only, transposed,
// projected, bounded, filtered, and stacked views. These adapters generally
// preserve the View interface while remapping coordinates so transformations can
// be chained without immediately materializing a new sheet.
//
// Formula support bridges the parser and value packages. EvalString parses OXML
// or ODS-style formulas and evaluates them against a Context. Evaluation returns
// value.Value instances, including spreadsheet-style error values for ordinary
// formula errors such as invalid references, unknown functions, or division by
// zero.
package grid
