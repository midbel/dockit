// Package eval executes Dockit scripts.
//
// Engine is the main entry point. It scans and parses a script, extracts script
// configuration, constructs an EngineContext, and evaluates the resulting AST
// against a formula/env environment. The engine registers file loaders for CSV,
// XLSX, ODS, log, JSON, JSON5, and XML input by default.
//
// EngineContext implements value.Context for formulas. It resolves variables,
// tracks the current default file or view, reads cells and ranges, applies
// mutations such as insert and remove, formats printed output, and coordinates
// import/export behavior.
//
// The evaluator uses the visitor interfaces from formula/parse. It supports
// scalar, array, range, view, and object operations, including vectorized calls
// and special forms such as inspect and kindof. Several parsed language forms
// are still placeholders or incomplete, so this package should be considered
// active implementation code rather than a stable embeddable scripting API.
package eval
