package parse

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
)

type AddressContext int8

const (
	DefaultAddressContext AddressContext = iota
	DottedAddressContext
	BracketAddressContext
)

type DialectRules interface {
	ParseAddress(*Parser, AddressContext) (Expr, error)
	ParseQualifiedAddress(*Parser, Expr) (Expr, error)
	ParseRangeAddress(*Parser, Expr) (Expr, error)

	Separator() op.Op
}

func FormulaGrammar() *Grammar {
	g := NewGrammar("formula")

	g.terminators = []op.Op{op.EOF}

	g.RegisterPrefix(op.Cell, func(p *Parser) (Expr, error) {
		return p.dialect.ParseAddress(p, DefaultAddressContext)
	})
	g.RegisterPrefix(op.BegAddr, func(p *Parser) (Expr, error) {
		return p.dialect.ParseAddress(p, BracketAddressContext)
	})
	g.RegisterPrefix(op.SheetRef, func(p *Parser) (Expr, error) {
		return p.dialect.ParseAddress(p, DottedAddressContext)
	})
	g.RegisterPrefix(op.Number, parseNumber)
	g.RegisterPrefix(op.Literal, parseLiteral)
	g.RegisterPrefix(op.Sub, parseUnary)
	g.RegisterPrefix(op.Add, parseUnary)
	g.RegisterPrefix(op.Ident, parseIdentifier)
	g.RegisterPrefix(op.BegGrp, parseGroup)

	g.RegisterPostfix(op.SheetRef, func(p *Parser, expr Expr) (Expr, error) {
		return p.dialect.ParseQualifiedAddress(p, expr)
	})
	g.RegisterPostfix(op.BegGrp, parseCall)
	g.RegisterPostfix(op.Percent, parsePercent)

	g.RegisterInfix(op.RangeRef, func(p *Parser, expr Expr) (Expr, error) {
		return p.dialect.ParseRangeAddress(p, expr)
	})
	g.RegisterInfix(op.Add, parseBinary)
	g.RegisterInfix(op.Sub, parseBinary)
	g.RegisterInfix(op.Mul, parseBinary)
	g.RegisterInfix(op.Div, parseBinary)
	g.RegisterInfix(op.Concat, parseBinary)
	g.RegisterInfix(op.Pow, parseBinary)
	g.RegisterInfix(op.Eq, parseBinary)
	g.RegisterInfix(op.Ne, parseBinary)
	g.RegisterInfix(op.Lt, parseBinary)
	g.RegisterInfix(op.Le, parseBinary)
	g.RegisterInfix(op.Gt, parseBinary)
	g.RegisterInfix(op.Ge, parseBinary)

	return g
}

func ScriptGrammar() *Grammar {
	g := FormulaGrammar()
	g.name = "script"

	g.terminators = []op.Op{op.EOF, op.Eol, op.Semi}

	g.RegisterPrefix(op.Eq, parseDeferred)
	g.RegisterPrefix(op.Ident, parseIdentifier)
	g.RegisterPrefix(op.Column, parseColumn)
	g.RegisterPrefix(op.BegProp, parseSlicePrefix)
	g.RegisterPrefix(op.BegGrp, parseGroup)
	g.RegisterPrefix(op.Special, parseSpecialAccessPrefix)

	g.RegisterPostfix(op.Dot, parseAccess)
	g.RegisterPostfix(op.Special, parseSpecialAccess)
	g.RegisterPostfix(op.BegProp, parseSlice)

	g.RegisterInfix(op.Union, parseBinary)
	g.RegisterInfix(op.Assign, parseAssignment)
	g.RegisterInfix(op.AddAssign, parseAssignment)
	g.RegisterInfix(op.SubAssign, parseAssignment)
	g.RegisterInfix(op.MulAssign, parseAssignment)
	g.RegisterInfix(op.PowAssign, parseAssignment)
	g.RegisterInfix(op.DivAssign, parseAssignment)
	g.RegisterInfix(op.ConcatAssign, parseAssignment)

	g.RegisterPrefixKeyword(kwAssert, parseAssert)
	g.RegisterPrefixKeyword(kwUse, parseUse)
	g.RegisterPrefixKeyword(kwImport, parseImport)
	g.RegisterPrefixKeyword(kwPrint, parsePrint)
	g.RegisterPrefixKeyword(kwExport, parseExport)
	// g.RegisterPrefixKeyword(kwInclude, parseInclude)
	// g.RegisterPrefixKeyword(kwMacro, parseMacro)

	return g
}

func LambdaGrammar() *Grammar {
	g := FormulaGrammar()
	g.name = "lambda"
	g.scope = GrammarIsolated
	g.RegisterPostfix(op.BegProp, parseSlice)
	return g
}

func SliceGrammar() *Grammar {
	g := NewGrammar("slice")
	g.scope = GrammarIsolated
	g.bindings[op.Semi] = powList

	g.RegisterPrefix(op.Cell, func(p *Parser) (Expr, error) {
		return p.dialect.ParseAddress(p, DefaultAddressContext)
	})
	g.RegisterPrefix(op.Ident, parseIdentifier)
	g.RegisterPrefix(op.Column, parseColumn)
	g.RegisterPrefix(op.Number, parseNumber)
	g.RegisterPrefix(op.Literal, parseLiteral)
	g.RegisterPrefix(op.RangeRef, parseOpenSelectedColumns)
	g.RegisterPrefix(op.BegGrp, parseGroup)
	g.RegisterPrefix(op.Not, parseNot)

	g.RegisterInfix(op.BegGrp, parseCall)

	g.RegisterInfix(op.RangeRef, parseRangeColumns)
	g.RegisterInfix(op.Semi, parseSelectedColumns)

	g.RegisterInfix(op.Eq, parseBinary)
	g.RegisterInfix(op.Ne, parseBinary)
	g.RegisterInfix(op.Lt, parseBinary)
	g.RegisterInfix(op.Le, parseBinary)
	g.RegisterInfix(op.Gt, parseBinary)
	g.RegisterInfix(op.Ge, parseBinary)
	g.RegisterInfix(op.And, parseAnd)
	g.RegisterInfix(op.Or, parseOr)

	return g
}

type ConfigEntry struct {
	Path  []string
	Value any
}

type Mode string

const (
	ModeScript   = "script"
	ModePipeline = "pipeline"
	ModeCube     = "cube"
	ModeCommand  = "command"
	ModeFormula  = "formula"
)

type Parser struct {
	scan Scanner
	curr Token
	peek Token

	mode Mode

	stack   *GrammarStack
	dialect DialectRules
}

func ParseFormula(str string) (Expr, error) {
	return ParseOxmlFormula(str)
}

func ParseOxmlFormula(str string) (Expr, error) {
	scan, err := ScanOxml(strings.NewReader(str))
	if err != nil {
		return nil, err
	}
	p, err := NewParser(scan)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

func ParseOdsFormula(str string) (Expr, error) {
	scan, err := ScanOds(strings.NewReader(str))
	if err != nil {
		return nil, err
	}
	p, err := NewParser(scan)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

func NewParser(scan Scanner) (*Parser, error) {
	p := Parser{
		scan:  scan,
		stack: new(GrammarStack),
		mode:  ModeScript,
	}
	switch scan.Type() {
	case TypeOds:
		p.dialect = odsDialectRules{}
	case TypeOxml:
		p.dialect = oxmlDialectRules{}
	default:
		return nil, fmt.Errorf("unsupported dialect")
	}
	p.next()
	p.next()
	if scan.Script() {
		mode, err := p.parseMode()
		if err != nil {
			return nil, err
		}
		switch mode {
		case "", ModeScript:
			mode = ModeScript
			p.pushGrammar(ScriptGrammar())
		default:
			return nil, fmt.Errorf("%s: mode not yet supported", mode)
		}
		p.mode = mode
	} else {
		p.pushGrammar(FormulaGrammar())
	}
	return &p, nil
}

func parseExprFromString(str string) (Expr, error) {
	scan, err := ScanScript(strings.NewReader(str))
	if err != nil {
		return nil, err
	}
	p, err := NewParser(scan)
	if err != nil {
		return nil, err
	}
	x, err := p.Parse()
	if err != nil {
		return nil, err
	}
	s, ok := x.(Script)
	if !ok || len(s.Body) != 1 {
		fmt.Println(str)
		return nil, fmt.Errorf("invalid script string")
	}
	return s.Body[0], nil
}

func (p *Parser) Parse() (Expr, error) {
	if !p.scan.Script() {
		return p.parseFormula()
	}
	return p.parseScript()
}

func (p *Parser) ExtractConfigEntries() ([]ConfigEntry, error) {
	p.skipEOL()
	var entries []ConfigEntry
	for p.is(op.Pragma) {
		var entry ConfigEntry
		for !p.done() {
			p.next()
			if p.is(op.Dot) {
				continue
			}
			if p.is(op.Assign) {
				break
			}
			if !p.is(op.Ident) && !p.is(op.Keyword) {
				return nil, fmt.Errorf("identifier expected")
			}
			entry.Path = append(entry.Path, p.currentLiteral())
		}
		if !p.is(op.Assign) {
			return nil, fmt.Errorf("assignment operator is expected")
		}
		p.next()
		entry.Value = p.value()
		p.next()
		if !p.isEOL() {
			return nil, p.makeError("eol expected")
		}
		p.skipEOL()
		entries = append(entries, entry)
	}
	return entries, nil
}

func (p *Parser) Mode() Mode {
	return p.mode
}

func (p *Parser) ParseNext() (Expr, error) {
	if p.done() {
		return nil, io.EOF
	}
	p.skipTerminator()
	p.skipComment()
	return p.parse(powLowest)
}

func (p *Parser) parseMode() (Mode, error) {
	if !p.is(op.Directive) {
		return ModeScript, nil
	}
	if p.currentLiteral() == "" {
		return "", fmt.Errorf("empty directive")
	}
	mode := p.currentLiteral()
	p.next()
	p.skipEOL()
	return Mode(mode), nil
}

func (p *Parser) parseFormula() (Expr, error) {
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, p.makeError("invalid formula given")
	}
	return expr, nil
}

func (p *Parser) parseScript() (Expr, error) {
	var (
		script Script
		err    error
	)
	script.Includes, err = p.parseIncludes()
	if err != nil {
		return nil, err
	}
	script.Body, err = p.parseBody()
	if err != nil {
		return nil, err
	}
	return script, nil
}

func (p *Parser) parseBody() ([]Expr, error) {
	var list []Expr
	for {
		p.skipComment()
		if p.done() {
			break
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
		if !p.isTerminator() {
			return nil, p.expectedEOL()
		}
		p.skipTerminator()
	}
	return list, nil
}

func (p *Parser) parseIncludes() ([]Expr, error) {
	var list []Expr
	for {
		p.skipComment()
		if p.done() {
			break
		}
		if p.is(op.Keyword) && p.currentLiteral() == kwInclude {
			expr, err := parseInclude(p)
			if err != nil {
				return nil, err
			}
			list = append(list, expr)
		} else {
			break
		}
		if !p.isTerminator() {
			return nil, p.expectedEOL()
		}
		p.skipTerminator()
	}
	p.currGrammar().UnregisterPrefixKeyword(kwInclude)
	return list, nil
}

func (p *Parser) parse(pow int) (Expr, error) {
	fn, err := p.prefix()
	if err != nil {
		return nil, err
	}
	left, err := fn(p)
	if err != nil {
		return nil, err
	}
	for {
		fn, err := p.postfix()
		if err != nil {
			break
		}
		left, err = fn(p, left)
		if err != nil {
			return nil, err
		}
	}
	for !p.isTerminator() && pow < p.pow(p.curr.Type) {
		fn, err := p.infix()
		if err != nil {
			return nil, err
		}
		left, err = fn(p, left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) value() any {
	switch str := p.currentLiteral(); p.curr.Type {
	case op.Number:
		f, _ := strconv.ParseFloat(str, 64)
		return f
	case op.Ident:
		if str == "true" {
			return true
		}
		return false
	case op.Literal:
		return str
	default:
		return nil
	}
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) done() bool {
	return p.is(op.EOF)
}

func (p *Parser) is(kind op.Op) bool {
	return p.curr.Type == kind
}

func (p *Parser) isTerminator() bool {
	return p.currGrammar().IsTerminator(p.curr)
}

func (p *Parser) isEOL() bool {
	return p.is(op.Eol)
}

func (p *Parser) skipEOL() {
	for !p.done() && p.is(op.Eol) {
		p.next()
	}
}

func (p *Parser) skipTerminator() {
	for p.isTerminator() && !p.done() {
		p.next()
	}
}

func (p *Parser) skipComment() {
	for p.is(op.Comment) {
		p.next()
	}
}

func (p *Parser) currentLiteral() string {
	return p.curr.Literal
}

func (p *Parser) pow(kind op.Op) int {
	return p.currGrammar().Pow(kind)
}

func (p *Parser) prefix() (PrefixFunc, error) {
	return p.currGrammar().Prefix(p.curr)
}

func (p *Parser) postfix() (InfixFunc, error) {
	return p.currGrammar().Postfix(p.curr)
}

func (p *Parser) infix() (InfixFunc, error) {
	return p.currGrammar().Infix(p.curr)
}

func (p *Parser) pushGrammar(g *Grammar) error {
	return p.stack.Push(g)
}

func (p *Parser) popGrammar() {
	p.stack.Pop()
}

func (p *Parser) currGrammar() *Grammar {
	return p.stack.Top()
}

func (p *Parser) makeError(msg string) error {
	return fmt.Errorf("(%s) %s: %s", p.currGrammar().Context(), p.curr.Position, msg)
}

func (p *Parser) expectedEOL() error {
	return p.makeError("end of line expected")
}

func (p *Parser) expectedIdent() error {
	return p.makeError("identifier expected")
}

func parseCall(p *Parser, expr Expr) (Expr, error) {
	if _, ok := expr.(Identifier); !ok {
		return nil, fmt.Errorf("identifier expected")
	}
	p.next()
	var args []Expr
	for !p.done() && !p.is(op.EndGrp) {
		p.skipEOL()
		arg, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		switch p.curr.Type {
		case op.Eol:
			p.skipEOL()
			if !p.is(op.EndGrp) {
				return nil, p.makeError("expected ')' after last argument")
			}
		case p.dialect.Separator():
			p.next()
			p.skipEOL()
			if p.is(op.EndGrp) {
				return nil, p.makeError("unexpected ')' after ','")
			}
		case op.EndGrp:
		default:
			return nil, p.makeError("unexpected character in function call")
		}
		args = append(args, arg)
	}
	if !p.is(op.EndGrp) {
		return nil, p.makeError("unexpected character in function call")
	}
	p.next()
	return NewCall(expr, args), nil
}

func parseBinary(p *Parser, left Expr) (Expr, error) {
	oper := p.curr.Type
	p.next()
	right, err := p.parse(p.pow(oper))
	if err != nil {
		return nil, err
	}
	return NewBinary(left, right, oper), nil
}

func parseUnary(p *Parser) (Expr, error) {
	oper := p.curr.Type
	p.next()
	right, err := p.parse(powUnary)
	if err != nil {
		return nil, err
	}
	return NewUnary(right, oper), nil
}

func parsePercent(p *Parser, expr Expr) (Expr, error) {
	expr = NewPostfix(expr, p.curr.Type)
	p.next()
	return expr, nil
}

func parseGroup(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if !p.is(op.EndGrp) {
		return nil, p.makeError("missing ')' at end of expression")
	}
	p.next()
	return expr, nil
}

func parseNumber(p *Parser) (Expr, error) {
	defer p.next()

	x, err := strconv.ParseFloat(p.currentLiteral(), 64)
	if err != nil {
		return nil, err
	}
	return NewNumber(x), nil
}

func parseLiteral(p *Parser) (Expr, error) {
	lit := p.currentLiteral()
	p.next()
	if strings.Index(lit, "${") < 0 {
		return NewLiteral(lit), nil
	}
	var (
		offset int
		list   []Expr
	)
	for len(lit) > 0 {
		ix := strings.Index(lit[offset:], "${")
		if ix < 0 {
			break
		}
		list = append(list, NewLiteral(lit[offset:offset+ix]))
		offset += ix + 2
		if ix = strings.Index(lit[offset:], "}"); ix <= 0 {
			return nil, p.makeError("invalid template string")
		}
		expr, err := parseExprFromString(lit[offset : offset+ix])
		if err != nil {
			return nil, err
		}
		list = append(list, expr)
		offset += ix + 1
	}
	if len(lit[offset:]) > 0 {
		list = append(list, NewLiteral(lit[offset:]))
	}
	return NewTemplate(list), nil
}

func parseIdentifier(p *Parser) (Expr, error) {
	id := NewIdentifier(p.currentLiteral())
	p.next()
	return id, nil
}

func parseRangeAddress(p *Parser, left Expr) (Expr, error) {
	p.next()

	addr, err := parseAddress(p)
	if err != nil {
		return nil, err
	}

	var start CellAddr
	switch a := left.(type) {
	case CellAddr:
		start = a
	case Identifier:
		addr, err := parseCellAddr(a.Ident())
		if err != nil {
			return nil, err
		}
		start = addr
	case Number:
		addr, err := parseCellAddr(a.String())
		if err != nil {
			return nil, err
		}
		start = addr
	default:
		return nil, fmt.Errorf("range: address/identfier/number expected")
	}
	end, ok := addr.(CellAddr)
	if !ok {
		return nil, p.makeError("range (right): address expected")
	}

	return NewRangeAddr(start, end), nil
}

func parseQualifiedAddress(p *Parser, left Expr) (Expr, error) {
	p.next()
	right, err := p.parse(powSheet)
	if err != nil {
		return nil, err
	}
	switch r := right.(type) {
	case CellAddr:
	case RangeAddr:
	case Identifier:
		addr, err := parseCellAddr(r.Ident())
		if err != nil {
			return nil, err
		}
		right = addr
	default:
		return nil, p.makeError("address/range expected")
	}
	return NewCellAccess(left, right), nil
}

func parseColumn(p *Parser) (Expr, error) {
	addr, err := parseColumnAddr(p.currentLiteral())
	if err != nil {
		return nil, err
	}
	p.next()
	return addr, nil
}

func parseAddress(p *Parser) (Expr, error) {
	addr, err := parseCellAddr(p.currentLiteral())
	if err != nil {
		return nil, err
	}
	p.next()
	return addr, nil
}

func parseDeferred(p *Parser) (Expr, error) {
	g := LambdaGrammar()
	if err := p.pushGrammar(g); err != nil {
		return nil, err
	}
	defer p.popGrammar()

	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	return NewDeferred(expr), nil
}

func parseAccess(p *Parser, left Expr) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	defer p.next()
	prop := NewIdentifier(p.currentLiteral())
	return NewAccess(left, prop), nil
}

func parseSpecialAccessPrefix(p *Parser) (Expr, error) {
	return parseSpecialAccess(p, nil)
}

func parseSpecialAccess(p *Parser, left Expr) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	defer p.next()
	prop := NewIdentifier(p.currentLiteral())
	return NewSpecial(left, prop), nil
}

func parseAssignment(p *Parser, left Expr) (Expr, error) {
	oper := p.curr.Type
	p.next()

	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	switch oper {
	case op.Assign:
	case op.AddAssign:
		expr = NewBinary(left, expr, op.Add)
	case op.SubAssign:
		expr = NewBinary(left, expr, op.Sub)
	case op.MulAssign:
		expr = NewBinary(left, expr, op.Mul)
	case op.DivAssign:
		expr = NewBinary(left, expr, op.Div)
	case op.PowAssign:
		expr = NewBinary(left, expr, op.Pow)
	case op.ConcatAssign:
		expr = NewBinary(left, expr, op.Concat)
	default:
	}
	return NewAssignment(left, expr), nil
}

func parsePrint(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	stmt := PrintRef{
		expr: expr,
	}
	if p.is(op.Literal) {
		stmt.pattern = p.currentLiteral()
		p.next()
	}
	return stmt, nil
}

func parseExport(p *Parser) (Expr, error) {
	p.next()
	var (
		stmt ExportRef
		err  error
	)
	if stmt.expr, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	if p.is(op.Keyword) && p.currentLiteral() == kwUsing {
		p.next()
		if !p.is(op.Ident) {
			return nil, p.makeError("identifier expected")
		}
		stmt.format = p.currentLiteral()
		p.next()
	}
	if p.is(op.Keyword) && p.currentLiteral() == kwWith {
		p.next()
		if p.is(op.Literal) {
			stmt.specifier = p.currentLiteral()
			p.next()
		} else if p.is(op.BegGrp) {
			opts, err := parseKeyValuePairs(p)
			if err != nil {
				return nil, err
			}
			stmt.options = opts
		} else {
			return nil, p.makeError("literal or key/value pair expected")
		}
	}
	if !p.is(op.Keyword) && p.currentLiteral() != kwTo {
		return nil, p.makeError("keyword 'to' expected")
	}
	p.next()
	if !p.is(op.Literal) {
		return nil, p.makeError("literal expected")
	}
	stmt.file = p.currentLiteral()
	p.next()
	return stmt, nil
}

func parseUse(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	stmt := UseRef{
		ident: p.currentLiteral(),
	}
	p.next()
	ro, err := parseReadonly(p)
	if err != nil {
		return nil, err
	}
	stmt.readOnly = ro
	return stmt, nil
}

func parseKeyValuePairs(p *Parser) (map[string]any, error) {
	p.next()
	kvs := make(map[string]any)
	for !p.done() && !p.is(op.EndGrp) {
		p.skipEOL()
		if !p.is(op.Ident) && !p.is(op.Literal) {
			return nil, p.makeError("only identifier or literal allowed as key")
		}
		key := p.currentLiteral()
		p.next()
		if !p.is(op.Assign) {
			return nil, p.makeError("assignment operator expected between key/value")
		}
		p.next()
		kvs[key] = p.currentLiteral()
		p.next()
		switch {
		case p.is(op.Eol):
			p.skipEOL()
			if !p.is(op.EndGrp) {
				return nil, p.makeError("expected ')' after last property")
			}
		case p.is(op.Comma):
			p.next()
			p.skipEOL()
			if p.is(op.EndGrp) {
				return nil, p.makeError("unexpected ')' after ','")
			}
		case p.is(op.EndGrp):
		default:
			return nil, p.makeError("')' or ',' expected")
		}
	}
	return kvs, nil
}

func parseAssert(p *Parser) (Expr, error) {
	p.next()
	var mode AssertType
	if p.is(op.Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(op.Ident) {
			return nil, p.expectedIdent()
		}
		switch p.currentLiteral() {
		case "fail":
			mode = AssertFail
		case "ignore":
			mode = AssertIgnore
		case "warn":
			mode = AssertWarn
		default:
			return nil, fmt.Errorf("%s: unknown assertion mode", p.currentLiteral())
		}
		p.next()
	} else {
		mode = AssertUnknown
	}
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	var msg string
	if p.is(op.Keyword) && p.currentLiteral() == kwElse {
		p.next()
		if !p.is(op.Literal) {
			msg := fmt.Sprintf("literal expected instead of %s", p.curr)
			return nil, p.makeError(msg)
		}
		msg = p.currentLiteral()
		p.next()
	}
	return NewAssert(expr, msg, mode), nil
}

func parseMacro(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	var (
		name = p.currentLiteral()
		args []Expr
		body []Expr
	)
	p.next()
	if !p.is(op.BegGrp) {
		return nil, p.makeError("expected '(' before arguments list")
	}
	p.next()
	for !p.done() && !p.is(op.EndGrp) {
		p.skipEOL()
		if !p.is(op.Ident) {
			return nil, p.expectedIdent()
		}
		args = append(args, NewIdentifier(p.currentLiteral()))
		p.next()

		switch p.curr.Type {
		case op.Eol:
			p.skipEOL()
			if !p.is(op.EndGrp) {
				return nil, p.makeError("expected ')' after last argument")
			}
		case op.Comma:
			p.next()
			p.skipEOL()
			if p.is(op.EndGrp) {
				return nil, p.makeError("unexpected ')' after ','")
			}
		case op.EndGrp:
		default:
			return nil, p.makeError("unexpected character in function call")
		}
	}
	if !p.is(op.EndGrp) {
		return nil, p.makeError("unexpected character in function call")
	}
	p.next()
	if !p.isTerminator() {
		return nil, p.expectedEOL()
	}
	p.skipTerminator()

	for !p.done() && !(p.is(op.Keyword) && p.currentLiteral() == kwEnd) {
		p.skipComment()
		if p.done() {
			break
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		body = append(body, e)
		if !p.isTerminator() {
			return nil, p.expectedEOL()
		}
		p.skipTerminator()
	}
	if !p.is(op.Keyword) && p.currentLiteral() != kwEnd {
		return nil, p.makeError("end keyword expected at end of macro")
	}
	p.next()
	return NewMacro(name, args, body), nil
}

func parseInclude(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Literal) {
		msg := fmt.Sprintf("literal expected instead of %s", p.curr)
		return nil, p.makeError(msg)
	}
	var (
		file  = p.currentLiteral()
		alias string
	)
	p.next()
	if p.is(op.Keyword) && p.currentLiteral() == "as" {
		p.next()
		if !p.is(op.Ident) {
			return nil, p.expectedIdent()
		}
		alias = p.currentLiteral()
		p.next()
	}
	return NewInclude(file, alias), nil
}

func parseImport(p *Parser) (Expr, error) {
	p.next()
	var stmt ImportFile
	if !p.is(op.Literal) {
		msg := fmt.Sprintf("literal expected instead of %s", p.curr)
		return nil, p.makeError(msg)
	}
	stmt.file = p.currentLiteral()
	p.next()
	if p.is(op.Keyword) && p.currentLiteral() == kwUsing {
		p.next()
		if !p.is(op.Ident) {
			msg := fmt.Sprintf("literal/identifier expected instead of %s", p.curr)
			return nil, p.makeError(msg)
		}
		stmt.format = p.currentLiteral()
		p.next()
		if p.is(op.Keyword) && p.currentLiteral() == kwWith {
			p.next()
			if p.is(op.BegGrp) {
				options, err := parseKeyValuePairs(p)
				if err != nil {
					return nil, err
				}
				stmt.options = options
			} else {
				stmt.specifier = p.currentLiteral()
			}
			p.next()
		}
	}
	if p.is(op.Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(op.Ident) {
			msg := fmt.Sprintf("literal/identifier expected instead of %s", p.curr)
			return nil, p.makeError(msg)
		}
		stmt.alias = p.currentLiteral()
		p.next()
	}
	if p.is(op.Keyword) && p.currentLiteral() == kwDefault {
		p.next()
		stmt.defaultFile = true
	}
	ro, err := parseReadonly(p)
	if err != nil {
		return nil, err
	}
	stmt.readOnly = ro
	return stmt, nil
}

func parseReadonly(p *Parser) (bool, error) {
	if !p.is(op.Keyword) {
		return false, nil
	}
	var ok bool
	switch kw := p.currentLiteral(); kw {
	case kwRo:
		ok = true
	case kwRw:
	default:
		msg := fmt.Sprintf("%s: unexpected keyword", kw)
		return ok, p.makeError(msg)
	}
	p.next()
	return ok, nil
}

func parseSlicePrefix(p *Parser) (Expr, error) {
	return parseSlice(p, nil)
}

func parseSlice(p *Parser, left Expr) (Expr, error) {
	g := SliceGrammar()
	if err := p.pushGrammar(g); err != nil {
		return nil, err
	}
	defer p.popGrammar()

	p.next()

	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if !p.is(op.EndProp) {
		return nil, p.makeError("expected ] at end of slice expression")
	}
	p.next()
	if _, ok := expr.(IntervalExpr); ok {
		expr = IntervalList{
			items: []Expr{expr},
		}
	}
	return NewSlice(left, expr), nil
}

func parseOpenSelectedColumns(p *Parser) (Expr, error) {
	p.next()
	var (
		expr IntervalExpr
		err  error
	)
	if !p.is(op.EndProp) && !p.is(op.Semi) {
		expr.to, err = p.parse(powList)
	}
	return expr, err
}

func parseRangeColumns(p *Parser, left Expr) (Expr, error) {
	p.next()
	var (
		right Expr
		err   error
	)
	if !p.is(op.EndProp) && !p.is(op.Semi) && !p.is(op.RangeRef) {
		right, err = p.parse(powRange)
	}
	if err != nil {
		return nil, err
	}

	leftAddr, leftAddrOk := left.(CellAddr)
	rightAddr, rightAddrOk := right.(CellAddr)

	if leftAddrOk && rightAddrOk {
		return NewRangeAddr(leftAddr, rightAddr), nil
	}
	if !leftAddrOk && (!rightAddrOk || right == nil) {
		var step Expr
		if p.is(op.RangeRef) {
			p.next()
			step, err = p.parse(powRange)
			if err != nil {
				return nil, err
			}
		}
		expr := NewInterval(left, right, step)
		return expr, nil
	}
	return nil, fmt.Errorf("use range address or columns range not both")
}

func parseSelectedColumns(p *Parser, left Expr) (Expr, error) {
	var list []Expr
	list = append(list, left)

	for !p.done() && p.is(op.Semi) {
		p.next()
		expr, err := p.parse(powList)
		if err != nil {
			return nil, err
		}
		list = append(list, expr)
	}
	return NewIntervalList(list), nil
}

func parseNot(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	return NewNot(expr), nil
}

func parseAnd(p *Parser, left Expr) (Expr, error) {
	p.next()
	right, err := p.parse(powLogical)
	if err != nil {
		return nil, err
	}
	return NewAnd(left, right), nil
}

func parseOr(p *Parser, left Expr) (Expr, error) {
	p.next()
	right, err := p.parse(powLogical)
	if err != nil {
		return nil, err
	}
	return NewOr(left, right), nil
}

type oxmlDialectRules struct{}

func (oxmlDialectRules) ParseAddress(p *Parser, ctx AddressContext) (Expr, error) {
	if ctx != DefaultAddressContext {
		return nil, fmt.Errorf("unsupported address type")
	}
	return parseAddress(p)
}

func (oxmlDialectRules) ParseQualifiedAddress(p *Parser, expr Expr) (Expr, error) {
	return parseQualifiedAddress(p, expr)
}

func (oxmlDialectRules) ParseRangeAddress(p *Parser, expr Expr) (Expr, error) {
	return parseRangeAddress(p, expr)
}

func (oxmlDialectRules) Separator() op.Op {
	return op.Comma
}

type odsDialectRules struct{}

func (odsDialectRules) ParseAddress(p *Parser, ctx AddressContext) (Expr, error) {
	switch ctx {
	case DefaultAddressContext:
		return parseAddress(p)
	case DottedAddressContext:
		p.next()
		return parseAddress(p)
	case BracketAddressContext:
		p.next()
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if !p.is(op.EndAddr) {
			return nil, fmt.Errorf("invalid address - missing ]")
		}
		p.next()
		return expr, nil
	default:
		return nil, fmt.Errorf("unsupported address type")
	}
}

func (odsDialectRules) ParseQualifiedAddress(p *Parser, expr Expr) (Expr, error) {
	return parseQualifiedAddress(p, expr)
}

func (odsDialectRules) ParseRangeAddress(p *Parser, expr Expr) (Expr, error) {
	return parseRangeAddress(p, expr)
}

func (odsDialectRules) Separator() op.Op {
	return op.Semi
}
