// Package value defines Dockit's runtime value model.
//
// The package is intentionally small and interface-oriented. Every runtime
// object implements Value, which reports a broad Kind and a more specific Type.
// ScalarValue covers primitive spreadsheet values such as Float, Text, Boolean,
// Date, Blank, and spreadsheet-style Error values. ArrayValue represents a
// rectangular collection of scalar values, while ObjectValue is used by richer
// script objects that expose named properties.
//
// Arithmetic and comparison helpers such as Add, Div, Eq, and Lt dispatch to
// optional methods implemented by concrete values and return spreadsheet errors
// instead of Go errors for ordinary formula failures. Cast helpers convert
// values between the primitive scalar domains used by formulas and built-ins.
//
// The package also defines the Context and Formula contracts used by grid and
// formula evaluators. Context is deliberately narrow: formulas can read a cell,
// read a range, or resolve a name without depending on a concrete workbook
// implementation.
package value
