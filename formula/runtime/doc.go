// Package types adapts grid and script objects to the value.Value interfaces
// used by Dockit's formula evaluator.
//
// The package wraps grid.File and grid.View implementations as object values,
// exposes ranges and arrays as runtime values, provides inspectable records for
// the inspect special form, and defines immutable environment helpers such as
// env and flag values. These wrappers are the bridge between the generic value
// model and workbook-specific behavior.
//
// View values provide object properties such as name, line count, column count,
// and read-only or locked state. They also coordinate mutations, broadcasting
// arrays into ranges, formula synchronization, projection, filtering, and
// bounded views. File values expose workbook-level sheet access and sheet
// management.
package runtime
