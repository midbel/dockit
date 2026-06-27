// Package formula documents the packages that implement Dockit's expression
// and scripting language.
//
// The executable pieces live in subpackages. Package parse scans and parses
// formulas and scripts into AST nodes. Package eval executes scripts against
// an environment and workbook context. Package types wraps grid files, views,
// ranges, and inspection records as runtime values. Package builtins registers
// script-level helper functions, while grid/builtins contains many
// spreadsheet-style formula functions. Package format renders parsed formula
// expressions back to OXML or ODS syntax.
//
// The formula tree is still evolving. Some syntax is parsed before the
// evaluator implements it completely, so callers should treat the package set
// as internal infrastructure for Dockit rather than a stable public language
// API.
package formula
