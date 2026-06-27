# Dockit

Dockit is an experimental command-line toolkit for working with tabular data
and spreadsheets from the terminal.

It is built around a small scripting language that treats CSV files,
spreadsheet workbooks, sheets, ranges, formulas, and derived views as things you
can compose directly. It can also ingest less tabular sources such as JSON, XML,
JSON5, and logs by passing them through loaders that reshape their data into the
rectangular view model Dockit works on. The goal is to make routine spreadsheet
work feel closer to a repeatable program than a sequence of manual clicks.

Dockit is currently early-stage software. The core ideas are present, the
package structure is taking shape, and several workflows are implemented, but
the language and file-format support are still evolving.

## Why Dockit Exists

Spreadsheets are often the place where real operational work happens: cleaning
CSV files, joining reference tables, reshaping columns, checking assumptions,
building reports, and handing data to someone who expects a workbook.

That work is valuable, but it is often hard to repeat. A manual spreadsheet
session can hide the decisions that produced the final file. A conventional
program can be repeatable, but it may feel too far away from the spreadsheet
model people already understand.

Dockit tries to sit between those worlds.

It keeps familiar spreadsheet ideas:

* cells and ranges
* formulas
* sheets and workbooks
* text, numbers, booleans, blanks, and spreadsheet-style errors

And it gives them a scriptable shape:

* import files, including structured and semi-structured sources
* assign cells, columns, ranges, and variables
* derive views
* filter, project, join, group, union, intersect, and compare data
* print or export results
* turn a transformation into something reviewable and repeatable

## Philosophy

Dockit is guided by a few small principles.

### Spreadsheets Are a Data Model

Dockit does not treat spreadsheets as a second-class export format. A workbook
is a meaningful structure: files contain sheets, sheets expose views, views have
bounds, cells can hold values or formulas, and formulas can be evaluated in a
context.

### Views Before Copies

Many transformations are represented as views over existing data. A projection,
bounded range, filtered result, transposed sheet, join, or union can behave like
a sheet without immediately becoming another physical sheet. This keeps the
model composable and lets transformations be layered.

### Spreadsheet Semantics Belong in the Terminal

Dockit uses spreadsheet-style values and errors where they make sense. Formula
failures can produce values such as `#VALUE!`, `#REF!`, and `#DIV/0!` instead of
forcing every formula problem into a Go error.

### Scripts Should Explain the Work

A Dockit script is meant to be read as a record of the transformation:

```dockit
import "sample.csv" as pjt default

D := lower(concatenate("github.com/", B, "/", A))

print @active
```

That script says what was loaded, what was derived, and what result was
produced. The aim is not to replace general-purpose programming; it is to make
spreadsheet-shaped work explicit.

### Be Honest About Incompleteness

Dockit is still under construction. Some syntax is parsed before it is fully
implemented by the evaluator. Some writer/export paths need more tests. The
project favors making these edges visible over pretending the tool is finished.

## What Dockit Can Do

Dockit currently has building blocks for:

* loading CSV, XLSX, ODS, log, JSON, JSON5, and XML data
* adapting structured and semi-structured inputs into sheet-like views
* reading and writing cells and ranges
* evaluating formulas
* deriving views from sheets
* assigning scalar, array, and range values
* printing values, arrays, and views
* joining and combining tabular views
* inspecting parsed scripts
* manipulating workbook sheets from the CLI

The command-line entry point is `dockit`.

Common commands include:

* `run` executes a Dockit script
* `dump` inspects a script AST
* `info` prints workbook information
* `print` prints sheet data
* `join`, `group`, `merge`, and related commands operate on tabular data
* `add`, `drop`, `rename`, `copy`, `lock`, and `unlock` manage sheets
* `builtins` lists available built-in functions

## Input Model

Dockit has one main working shape: a rectangular view. Once data has that shape,
it can be addressed with cells and ranges, transformed with formulas, combined
with other views, printed, or exported.

Some file formats already look like that:

* CSV and other delimited text files are flat tables
* XLSX/OpenXML workbooks contain sheets
* ODS workbooks contain sheets

Other sources need an interpretation step before Dockit can work with them.
JSON, JSON5, XML, and logs are handled by loaders that extract or reshape the
source into rows and columns. In scripts, this appears through the `using`
specifier and loader options.

Structured data can be queried into columns:

```dockit
import "lang.json" using json[[$.owner.name, $.languages.name, $.languages.star | 0]] as lang default
import "lang.xml" using xml[[$.owner.name, $.languages.language.name, $.languages.language.star | 0]] as lang default
```

Logs can be parsed with a pattern and exposed as rows:

```dockit
import "app.log" using log[["level=${level} user=${user} action=${action}"]] as events default
```

That adapter layer is part of Dockit's philosophy: the terminal workflow should
not only automate spreadsheets, but also bring nearby data sources into the same
spreadsheet-shaped workspace.

## A Small Example

Given a CSV file with project data, this script imports it, creates a GitHub URL
column, and prints a filtered view:

```dockit
import "sample.csv" as pjt default

D1 := lower(concatenate("github.com/", B1, "/", A1))
D2 := lower(concatenate("github.com/", B2, "/", A2))
D3 := lower(concatenate("github.com/", B3, "/", A3))
D4 := lower(concatenate("github.com/", B4, "/", A4))

print @active[C1 = 'run']
```

The same idea can be expressed with a whole-column assignment:

```dockit
import "sample.csv" as pjt default

D := lower(concatenate("github.com/", B, "/", A))

print @active
```

## Relational Examples

Dockit can treat sheets as relational views.

```dockit
import "users.csv" as usr
import "projects.csv" as pjt

view := join(usr@active, pjt@active, "A", "D")
print view[C:E;G;B]
```

Set-like operations are also available for compatible views:

```dockit
import "series1.csv" as fst
import "series2.csv" as snd

print "union"
print union(fst@active, snd@active)

print "intersect"
print intersect(fst@active, snd@active)

print "except"
print except(fst@active, snd@active)
```

## Project Layout

The repository is organized around a few main areas:

* `cmd/dockit` contains the CLI
* `value` defines the runtime value model
* `grid` defines files, views, cells, formulas, and evaluation context
* `gridx` contains relational view transformations
* `formula/parse` scans and parses formulas and scripts
* `formula/eval` executes Dockit scripts
* `formula/types` adapts files, views, ranges, and inspection values to the runtime model
* `ods` and `oxml` handle workbook formats
* `flat` and `csv` handle flat tabular inputs
* formula loaders adapt JSON, XML, JSON5, and log data into sheet-like views
* `examples` contains small scripts and sample data

For a language-oriented reference, see [book.md](book.md).

## Status

Dockit is not release-ready yet. Current rough edges include:

* the full test suite is not green
* the `cube` package currently does not compile
* export and writer behavior needs more coverage
* several parser features are ahead of evaluator support
* examples and documentation are still being filled in

The short version: Dockit is a working idea with useful pieces, not a polished
product.

## Development

Run focused tests while working on a package:

```sh
go test ./value
go test ./gridx
go test ./formula/... -run '^$'
```

The full test command is:

```sh
go test ./...
```

At the moment, the full command is expected to expose existing failures. Treat
those failures as part of the current project backlog, not as a surprise.

## License

Dockit is released under the terms of the MIT license. See [LICENSE](LICENSE).
