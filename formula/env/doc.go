// Package env provides the mutable name environment used by Dockit scripts.
//
// An Environment maps identifiers to runtime value.Value instances. Missing
// names resolve to value.ErrRef so undefined identifiers can flow through the
// formula error model. Define replaces existing bindings unless the existing
// value implements ImmutableValue and reports that it is immutable.
//
// The package is intentionally minimal; higher-level behavior such as default
// workbooks, active views, loaders, printers, and configuration lives in
// formula/eval.
package env
