// Package repr converts Dockit scripts into an inspectable AST representation.
//
// Inspect and InspectFile parse a script and return an Envelop containing
// metadata plus a tree of Node values. Each node records a stable id, a broad
// type such as statement or expression, a specific name, optional primitive
// value data, parameters, and child nodes.
//
// The representation is intended for debugging, CLI inspection, editor tooling,
// and documentation. It is derived from formula/parse visitors and mirrors the
// parser AST rather than evaluated runtime values.
package repr
