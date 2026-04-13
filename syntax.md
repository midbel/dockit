# Dockit DSL syntax

Dockit offers a small domain-specific language designed to manipulate tabular data (such as spreadsheets) in a concise and expressive way.

It combines concepts from **OpenXML/OpenFormula (Excel-like formulas)** with lightweight scripting capabilities, allowing both **data transformation** and **automation** in a single language.


## Basic expressions

### Comments

Use `#` to add comments only at beginning of line. Comments are ignored during execution.

```
# this is a comment
```

Currently, comments are only allowed at the beginning of a line.

### Literals

Strings can be defined using single (`'`) or double (`"`) quotes.

```
'string'
"string"
```

The second form, surrounded by `"`, allows for template string to be used.

Template strings provide a convenient way to embed values directly inside a string using placeholders, without requiring explicit string concatenation.

They are particularly useful for building dynamic text from variables, cells, or expressions.

```
age := 42
print "Total ${A10}"
print "Foobar is ${age}"
```

### Numbers

Dockit has support for classic number formats. Numbers are represented internally as decimal:

* Integers
* Floating-point numbers
* Scientific notation (implementation-dependent)

```
42
3.14
2e+11
```

### Dates

Dates are handled through built-in functions (e.g. `date`, `today`, `now`). 

They are internally treated as numeric or structured values depending on context.

### Boolean

Boolean values represent logical truth:

```
true
false
```

### Identifiers

Identifiers are used to name variables.

They must start with a letter and can only contain alphanumeric characters.

### Working with cells and ranges

#### Cell Addresses

Dockit uses Excel-style references:

```
A1
$A$1
$A1
A$1
```

#### Ranges

Select Multiple cells via address or full columns directly

```
A1:C100
A:C
```

### View and File

#### Properties of File

A file represents a workbook.

Available properties:

* sheets: number of sheets in the file
* readonly: flag indicating if file is accessible in read only mode
* protected/locked: file is locked
* active: active sheet of a file

Access properties using `@`:

```
file@active
```

Access a specific sheet with `.`:

```
file.sheet
```

#### Properties of View

A view represents a sheet or dataset.

Available properties:

* name: name of sheet
* lines: number of rows in a sheet
* columns: number of columns in a sheet
* cells: number of cells in a sheet
* readonly: flag indicating if view is accessible in read only mode
* protected/locked: file is locked

Access properties using `.`:

```
view.name
```

#### Creating view from other view

* Bounding: limited rectangular area on top of base view
* Projection: selection of specific columns on top of base view
* Filter: selection of rows based on a predicate

When creating a new view, the new view coordinates starts at position (1, 1) and ends at (rows, columns) where (rows, columns) depends of the source view

### Deferred expressions

Use `=` to indicate a formula that should be evaluated later (like Excel formulas). This is usefull when assigning formula(s) to cell or range

```
= (A1 * 100) / 42
```

### Unary expressions

Apply a sign:

```
-var
+var
```

### Binary expressions

#### Scalar expression

```
var1 + var2
var1 - var2
var1 * var2
var1 / var2
var1 ^ var2 # power
var1 & var2 # concatenation
```

#### Logical operators

var1 = var2
var1 <> var2
var1 > var2
var1 < var2
var1 >= var2
var1 <= var2

#### array/range expression

Apply operations element-wise:

```
a1:c100 + a1:c100
a1:c100 - a1:c100
a1:c100 * a1:c100
a1:c100 / a1:c100
a1:c100 & a1:c100
```

#### array - scalar expression

Apply a value to all cells in a range:

```
a1:c100 + 42
a1:c100 - 42
a1:c100 * 42
a1:c100 / 42
a1:c100 & 42
```

Only allows when the scalar is on the right of the expression.

### View expressions

Views let you combine datasets.

```
view1 & view2 - vertical concatenation
view1 | view2 - horizontal concatenation
```

### Slices

```
view1[<binary expr>]
view1[<selection>]
view1[<range>]
```

### Assigmnent

Assign values to variables or ranges:

```
var := value
A1:C100 := 1
A1:C100 := A1:C100 * 2

sheet!A1 := =sum(A1:C100)
```

### Compound assignment

```
var += value
var -= value
var *= value
var /= value
var ^= value
```

## Statements

### print

Display content of variable/cell/array/view

```
print <expr>
```

### import

Load external data

```
import <file> [using <format>] [with <specifier|options>] [as <alias>] [default] [ro|rw]

```

Supported format:

* xlsx
* ods
* csv (or similar)
* log

### use

Specify data to be used as default:

```
use <ident> [ro|rw]
```

default object can then be accessed without their full name

### export

Save results to a file:

```
export <expr> [using <format>] [with <specifier|options>] to <file>
```

### assert

Validate conditions:

```
assert as <fail|warn|ignore> <expr> [else <message>]
```

The valid modes are:
* fail: stop execution of current scriot
* warn: print on stdout given message
* ignore: do nothing

## builtins

### special script

* inspect
* kindof
* lock
* unlock
* mkaddr
* mkrange
* newfile
* newsheet
* join

### openxml/openformula

Dockit includes many built-ins inspired by Excel.

* and
* choose
* if
* iferror
* ifna
* ifs
* index
* match
* not
* or
* switch
* vlookup
* xor
* err
* na
* abs
* acos
* asin
* atan2
* average
* cos
* degress
* e
* exp
* int
* iseven
* isodd
* ln
* log10
* max
* median
* min
* mod
* mode
* pi
* power
* radians
* rand
* round
* rounddown
* roundup
* sign
* sin
* sqrt
* stdev
* sum
* tan
* var
* averageif
* count
* counta
* countif
* sumif
* type
* clean
* concatenate
* exact
* find
* left
* len
* lower
* mid
* proper
* replace
* rept
* right
* search
* substitue
* text
* textjoin
* trim
* upper
* value
* date
* datedif
* day
* hour
* minute
* month
* now
* second
* today
* weekday
* year
* yearday
* isblank
* iserror
* isna
* isnumber
* istext

## Script configuration

### startup

### directives