# The Dockit Book

Dockit is an experimental command-line tool and scripting language for working
with tabular data: CSV files, spreadsheet-like workbooks, sheets, ranges, and
formula expressions.

The language is still evolving. This book is a working reference for the parts
that are visible in the repository today, not a compatibility promise.

Dockit combines three ideas:

* spreadsheet-style values and formulas
* workbook and sheet operations
* script statements for importing, transforming, printing, checking, and
  exporting data

## Concepts

### Values

Dockit expressions evaluate to values from a small runtime model:

* numbers
* text
* booleans
* dates
* blanks
* arrays and ranges
* objects such as files and views
* spreadsheet errors such as `#VALUE!`, `#REF!`, and `#DIV/0!`

Errors are ordinary values in many formula paths. A failed spreadsheet operation
usually returns an error value, while a failed script operation can return a Go
error and stop execution.

### Files and Views

A file represents a workbook-like container. A view represents a sheet or a
derived rectangular dataset.

Views can be real sheets or adapters over other views:

* bounded views expose a rectangular slice of a source view
* projected views expose selected columns
* filtered views expose rows matching a predicate
* transposed views swap rows and columns
* stacked views combine several views

Derived views usually start their visible coordinates at `A1`, even when they
read from a different region of the source sheet.

## Basic Syntax

### Comments

Comments currently use `#` at the beginning of a line.

```dockit
# this is a comment
```

### Literals

Text can use single or double quotes.

```dockit
'plain text'
"template text"
```

Double-quoted strings can include template placeholders.

```dockit
age := 42
print "age = ${age}"
print "A10 = ${A10}"
```

Numbers include integers, decimals, and scientific notation where supported by
the scanner.

```dockit
42
3.14
2e+11
```

Booleans are written as:

```dockit
true
false
```

### Identifiers

Identifiers name variables, imported files, and views.

```dockit
total := 42
print total
```

## Cells and Ranges

Dockit uses spreadsheet-style references.

```dockit
A1
$A$1
$A1
A$1
```

Ranges select rectangular cell groups or whole columns.

```dockit
A1:C100
A:C
```

Sheet-qualified references use `!`.

```dockit
sheet1!A1
sheet2!A1:C10
```

## Expressions

### Deferred Formulas

A leading `=` creates a formula expression. This is useful when assigning a
formula to a cell or range for later evaluation.

```dockit
A1 := =sum(B1:B10)
```

### Arithmetic

```dockit
gross + tax
gross - discount
price * quantity
total / count
base ^ exponent
first & last
```

The `&` operator concatenates text.

### Comparisons

```dockit
a = b
a <> b
a > b
a < b
a >= b
a <= b
```

### Range Operations

Some operations can apply element by element to ranges or arrays.

```dockit
B2:B10 * 2
B2:B10 + C2:C10
```

Scalar-on-range operations are currently expected with the scalar on the right.

```dockit
B2:B10 + 1
```

## Files and Views

### Import

The `import` statement loads external data and binds it to an identifier.

```dockit
import "sample.csv" using csv[[comma]] as data default
```

Common options:

* `as <alias>` names the imported file
* `default` makes it the default workbook for unqualified references
* `ro` or `rw` controls read-only intent where supported

Supported loader formats in the engine include:

* `csv`
* `xlsx`
* `ods`
* `log`
* `json`
* `json5`
* `xml`

CSV delimiter specifiers include:

* `comma`
* `semicolon`
* `tab`
* `colon`
* `detect`

Structured imports use a query specifier.

```dockit
import "lang.json" using json[[$.owner.name, $.languages.name]] as lang default
import "lang.xml" using xml[[$.owner.name, $.languages.language.name]] as lang default
```

### Properties

Files and views expose properties.

```dockit
sheet := @active.name
rows := @active.lines
cols := @active.columns
count := data.sheets
```

The active view of the default file is available through `@active`.

### Slices

Slices derive a view or range from another view.

```dockit
data[A:C]
data[A1:C10]
data[stars > 10]
```

The exact predicate and selection grammar is still developing.

### Combining Views

Views can be combined with expression operators.

```dockit
left & right
left | right
```

The implementation also contains relational helpers for joins, grouping, union,
intersection, and difference in the `gridx` package.

## Assignment

Assign to variables:

```dockit
name := "dockit"
answer := 42
```

Assign to cells or ranges:

```dockit
A1 := 1
A1:C10 := 0
B2:B10 := B2:B10 * 2
sheet1!A1 := =sum(B1:B10)
```

Compound assignment is available for common operators.

```dockit
total += value
total -= value
total *= value
total /= value
total ^= value
```

## Statements

### print

Print a value, range, array, view, or object.

```dockit
print total
print @active
```

### use

Make a file or view the default object for unqualified references.

```dockit
use data
```

### export

Export writes an expression to a target file.

```dockit
export @active using ods to "out.ods"
```

Export support is still incomplete in the current implementation. Treat it as an
area under development and verify generated files carefully.

### assert

Assertions validate expectations while a script runs.

```dockit
assert total > 0
assert as warn total > 0 else "total should be positive"
assert as ignore total > 0
```

Modes:

* `fail` stops the script
* `warn` reports a failed assertion without aborting
* `ignore` suppresses the failure

### insert

Insert rows or columns into the active or named view.

```dockit
insert row into @active with 0
insert 2 rows after 1 into @active
insert column before first into @active with "tbd"
```

### remove

Remove rows or columns from the active or named view.

```dockit
remove row from @active
remove 2 rows after 1 from @active

remove first row from @active
remove last row from @active
```

### lock and unlock

Lock or unlock files and views that support those operations.

```dockit
lock data
unlock data
```

### sheet

The parser contains syntax for creating sheets, but this area is still under
active development.

```dockit
sheet "summary" using data as summary
```

## Built-ins

Dockit includes built-ins inspired by spreadsheet formulas. The exact list and
behavior should be checked against the implementation and tests.

Special script forms currently include:

* `inspect`
* `kindof`

Formula-style built-ins include logical, lookup, numeric, text, date/time, and
type-checking functions such as:

* `and`
* `or`
* `not`
* `if`
* `iferror`
* `index`
* `match`
* `vlookup`
* `sum`
* `average`
* `min`
* `max`
* `count`
* `counta`
* `sumif`
* `countif`
* `abs`
* `round`
* `sqrt`
* `text`
* `textjoin`
* `trim`
* `upper`
* `lower`
* `date`
* `today`
* `now`
* `year`
* `month`
* `day`
* `isblank`
* `iserror`
* `isnumber`
* `istext`

## Script Configuration

Scripts can contain configuration entries that are extracted before execution.
Configuration feeds runtime options such as context directory, printer behavior,
and formatting.

This part of the language needs more examples and stabilization.

## Current Limitations

The repository is not yet in release shape. Known rough edges include:

* the full test suite is not green
* the `cube` package currently does not compile
* several script features are parsed before they are fully implemented
* export and writer paths need stronger tests
* file format round-tripping needs more coverage

Use Dockit as an evolving toolkit for experimentation until those areas are
finished.
