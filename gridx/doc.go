// Package gridx provides relational transformations for grid views.
//
// The package builds new read-only grid.View implementations from existing
// views. Join creates an inner join using selected key columns. Union,
// Intersect, and Except perform set-like row operations on views with matching
// widths. Group collapses rows by key columns and appends aggregate columns.
//
// Transformations are lazy at the cell/row interface boundary: they keep enough
// index state to map output rows back to source views, then expose the result as
// another grid.View. The returned views are not generally writable; Sync returns
// grid.ErrSupported for operations that cannot be pushed back into source data.
//
// Keys are derived from scalar value type and string representation, so callers
// should normalize values before joining or grouping when textual and numeric
// forms must compare equal.
package gridx
