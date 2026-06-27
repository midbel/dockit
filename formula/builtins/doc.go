// Package builtins registers script-level built-ins for Dockit's formula
// evaluator.
//
// These built-ins complement the spreadsheet-style functions in grid/builtins.
// They work with formula/types values such as files, views, and ranges, and
// expose helpers for creating empty files or sheets, constructing addresses and
// ranges, generating sequences, and applying relational operations such as
// join, group, union, intersect, and except.
//
// Lookup returns the callable implementation for a registered name. List
// exposes the registered metadata used by the CLI and help output. Ordinary
// formula execution should call into this package through the evaluator rather
// than invoking built-ins directly.
package builtins
