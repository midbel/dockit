package eval

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/value"
)

func FormulaGrammar() *Grammar {
	g := NewGrammar("formula", ModeFormula)

	g.terminators = []op.Op{op.EOF}

	g.RegisterPrefix(op.Cell, parseAddress)
	g.RegisterPrefix(op.Number, parseNumber)
	g.RegisterPrefix(op.Literal, parseLiteral)
	g.RegisterPrefix(op.Sub, parseUnary)
	g.RegisterPrefix(op.Add, parseUnary)
	g.RegisterPrefix(op.BegGrp, parseGroup)

	g.RegisterPostfix(op.SheetRef, parseQualifiedAddress)
	g.RegisterPostfix(op.BegGrp, parseCall)
	g.RegisterPostfix(op.Percent, parsePercent)

	g.RegisterInfix(op.RangeRef, parseRangeAddress)
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
	g.mode = ModeScript

	g.terminators = []op.Op{op.EOF, op.Eol, op.Semi}

	g.RegisterPrefix(op.Eq, parseDeferred)
	g.RegisterPrefix(op.Ident, parseIdentifier)
	g.RegisterPrefix(op.Cell, parseAddress)
	g.RegisterPrefix(op.BegProp, parseSlicePrefix)
	g.RegisterPrefix(op.BegGrp, parseGroup)
	g.RegisterPrefix(op.SpreadRef, parseSpread)

	g.RegisterPostfix(op.Dot, parseAccess)
	g.RegisterPostfix(op.BegProp, parseSlice)
	g.RegisterPostfix(op.SheetRef, parseQualifiedAddress)

	g.RegisterInfix(op.Union, parseBinary)
	g.RegisterInfix(op.Assign, parseAssignment)
	g.RegisterInfix(op.AddAssign, parseAssignment)
	g.RegisterInfix(op.SubAssign, parseAssignment)
	g.RegisterInfix(op.MulAssign, parseAssignment)
	g.RegisterInfix(op.PowAssign, parseAssignment)
	g.RegisterInfix(op.DivAssign, parseAssignment)
	g.RegisterInfix(op.ConcatAssign, parseAssignment)

	g.RegisterPrefixKeyword(kwUse, parseUse)
	g.RegisterPrefixKeyword(kwImport, parseImport)
	g.RegisterPrefixKeyword(kwPrint, parsePrint)
	g.RegisterPrefixKeyword(kwSave, parseSave)
	g.RegisterPrefixKeyword(kwExport, parseExport)
	g.RegisterPrefixKeyword(kwWith, parseWith)
	g.RegisterPrefixKeyword(kwLock, parseLock)
	g.RegisterPrefixKeyword(kwUnlock, parseUnlock)
	g.RegisterPrefixKeyword(kwClear, parseClear)
	g.RegisterPrefixKeyword(kwPush, parsePush)
	g.RegisterPrefixKeyword(kwPop, parsePop)

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
	g := NewGrammar("slice", ModeScript)
	g.scope = GrammarIsolated
	g.bindings[op.Semi] = powList

	g.RegisterPrefix(op.Cell, parseAddress)
	g.RegisterPrefix(op.Ident, parseIdentifier)
	g.RegisterPrefix(op.Number, parseNumber)
	g.RegisterPrefix(op.Literal, parseLiteral)
	g.RegisterPrefix(op.RangeRef, parseOpenSelectedColumns)
	g.RegisterPrefix(op.BegGrp, parseGroup)
	g.RegisterPrefix(op.Not, parseNot)
	g.RegisterPrefix(op.SpreadRef, parseSpread)

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

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	stack *GrammarStack
}

func ParseFormula(str string) (value.Formula, error) {
	p := NewParser(FormulaGrammar())
	expr, err := p.ParseString(str)
	if err != nil {
		return nil, err
	}
	f := deferredFormula{
		expr: expr,
	}
	return &f, nil
}

func NewParser(g *Grammar) *Parser {
	var p Parser
	p.stack = new(GrammarStack)
	p.pushGrammar(g)
	return &p
}

func parseExprFromString(str string) (Expr, error) {
	p := NewParser(ScriptGrammar())
	x, err := p.ParseString(str)
	if err != nil {
		return nil, err
	}
	s, ok := x.(Script)
	if !ok || len(s.Body) != 1 {
		return nil, fmt.Errorf("invalid script string")
	}
	return s.Body[0], nil
}

func (p *Parser) ParseString(str string) (Expr, error) {
	return p.Parse(strings.NewReader(str))
}

func (p *Parser) Parse(r io.Reader) (Expr, error) {
	if err := p.Init(r); err != nil {
		return nil, err
	}
	if p.stack.Mode() == ModeFormula {
		return p.parseFormula()
	}
	return p.parseScript()
}

func (p *Parser) Attach(scan *Scanner) {
	p.scan = scan
	p.next()
	p.next()
	p.skipTerminator()
}

func (p *Parser) Init(r io.Reader) error {
	scan, err := Scan(r, p.stack.Mode())
	if err != nil {
		return err
	}
	p.scan = scan
	p.next()
	p.next()
	return nil
}

func (p *Parser) ParseNext() (Expr, error) {
	if p.done() {
		return nil, io.EOF
	}
	p.skipTerminator()
	p.skipComment()
	return p.parse(powLowest)
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
	var script Script
	for {
		p.skipComment()
		if p.done() {
			break
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		script.Body = append(script.Body, e)
		if !p.isTerminator() {
			return nil, p.expectedEOL()
		}
		p.skipTerminator()
	}
	return script, nil
}

func (p *Parser) parseUntil(ok func() bool) ([]Expr, error) {
	var body []Expr
	for !p.done() && ok() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		body = append(body, e)
	}
	return body, nil
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

func parseSpread(p *Parser) (Expr, error) {
	p.next()
	next, err := p.parse(powSpread)
	if err != nil {
		return nil, err
	}
	expr := spread{
		expr: next,
	}
	return expr, nil
}

func parseCall(p *Parser, expr Expr) (Expr, error) {
	p.next()
	var args []Expr
	for !p.done() && !p.is(op.EndGrp) {
		arg, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		switch p.curr.Type {
		case op.Comma:
			p.next()
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
	expr = postfix{
		expr: expr,
		op:   p.curr.Type,
	}
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

	start, ok := left.(cellAddr)
	if !ok {
		return nil, p.makeError("range: address expected")
	}
	end, ok := addr.(cellAddr)
	if !ok {
		return nil, p.makeError("range: address expected")
	}

	return NewRangeAddr(start, end), nil
}

func parseQualifiedAddress(p *Parser, left Expr) (Expr, error) {
	p.next()

	right, err := p.parse(powSheet)
	if err != nil {
		return nil, err
	}
	return NewQualifiedAddr(left, right), nil
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
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	e := deferred{
		expr: expr,
	}
	return e, nil
}

func parseAccess(p *Parser, left Expr) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	a := access{
		expr: left,
		prop: p.currentLiteral(),
	}
	p.next()
	return a, nil
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
	stmt := printRef{
		expr: expr,
	}
	if p.is(op.Literal) {
		stmt.pattern = p.currentLiteral()
		p.next()
	}
	return stmt, nil
}

func parseSave(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	stmt := saveRef{
		expr: expr,
	}
	return stmt, nil
}

func parseExport(p *Parser) (Expr, error) {
	p.next()
	var (
		stmt exportRef
		err  error
	)
	if stmt.expr, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	if !p.is(op.Keyword) && p.currentLiteral() != kwTo {
		return nil, p.makeError("keyword 'to' expected")
	}
	p.next()
	if stmt.file, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	if p.is(op.Keyword) && p.currentLiteral() == kwAs {
		p.next()
		stmt.format, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil
}

func parseUse(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	stmt := useRef{
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

func parseKeyValuePairs(p *Parser) (map[string]string, error) {
	p.next()
	kvs := make(map[string]string)
	for !p.done() && !p.is(op.EndGrp) {
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
		case p.is(op.Comma):
			p.next()
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

func parseImport(p *Parser) (Expr, error) {
	p.next()
	var stmt importFile
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

func parseClear(p *Parser) (Expr, error) {
	return nil, nil
}

func parsePush(p *Parser) (Expr, error) {
	return nil, nil
}

func parsePop(p *Parser) (Expr, error) {
	return nil, nil
}

func parseLock(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	stmt := lockRef{
		ident: p.currentLiteral(),
	}
	p.next()
	return stmt, nil
}

func parseUnlock(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.expectedIdent()
	}
	stmt := unlockRef{
		ident: p.currentLiteral(),
	}
	p.next()
	return stmt, nil
}

func parseWith(p *Parser) (Expr, error) {
	return nil, nil
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
	if _, ok := expr.(exprRange); ok {
		rg, err := getColumnsRangeFromExpr(expr)
		if err != nil {
			return nil, err
		}
		expr = columnsSlice{
			columns: []columnsRange{rg},
		}
	}
	s := slice{
		view: left,
		expr: expr,
	}
	return s, nil
}

func parseColumnExpr(expr Expr) (int, error) {
	e, ok := expr.(identifier)
	if !ok {
		return 0, fmt.Errorf("columns identifier expected")
	}
	ix, size := parseIndex(e.name)
	if size != len(e.name) {
		return 0, fmt.Errorf("invalid column index")
	}
	return int(ix), nil
}

func getColumnsRangeFromExpr(expr Expr) (columnsRange, error) {
	var (
		crg columnsRange
		err error
	)
	if _, ok := expr.(rangeAddr); ok {
		return crg, fmt.Errorf("address range not allowed in selection list")
	}
	switch expr := expr.(type) {
	case identifier:
		crg.from, err = parseColumnExpr(expr)
		crg.to = crg.from
	case exprRange:
		if expr.from != nil {
			crg.from, err = parseColumnExpr(expr.from)
			if err != nil {
				break
			}
		}
		if expr.to != nil {
			crg.to, err = parseColumnExpr(expr.to)
			if err != nil {
				break
			}
		}
		if n, ok := expr.step.(number); ok {
			crg.step = int(n.value)
		}
	default:
		return crg, fmt.Errorf("invalid columns selector")
	}
	return crg, err
}

func parseOpenSelectedColumns(p *Parser) (Expr, error) {
	p.next()
	var (
		expr exprRange
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
	if !p.is(op.EndProp) && !p.is(op.Semi) {
		right, err = p.parse(powRange)
	}
	if err != nil {
		return nil, err
	}

	leftAddr, leftAddrOk := left.(cellAddr)
	rightAddr, rightAddrOk := right.(cellAddr)

	if leftAddrOk && rightAddrOk {
		expr := rangeAddr{
			startAddr: leftAddr,
			endAddr:   rightAddr,
		}
		return expr, nil
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
		expr := exprRange{
			from: left,
			to:   right,
			step: step,
		}
		return expr, nil
	}
	return nil, fmt.Errorf("use range address or columns range not both")
}

func parseSelectedColumns(p *Parser, left Expr) (Expr, error) {
	var cs columnsSlice

	crg, err := getColumnsRangeFromExpr(left)
	if err != nil {
		return nil, err
	}
	cs.columns = append(cs.columns, crg)
	for !p.done() && p.is(op.Semi) {
		p.next()
		expr, err := p.parse(powList)
		if err != nil {
			return nil, err
		}
		crg, err := getColumnsRangeFromExpr(expr)
		if err != nil {
			return nil, err
		}
		cs.columns = append(cs.columns, crg)
	}
	return cs, nil
}

func parseFilterRows(p *Parser, left Expr) (Expr, error) {
	expr, err := parseBinary(p, left)
	if err != nil {
		return nil, err
	}
	fs := filterSlice{
		expr: expr,
	}
	return fs, nil
}

func parseNot(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	ret := not{
		expr: expr,
	}
	return ret, nil
}

func parseAnd(p *Parser, left Expr) (Expr, error) {
	p.next()
	right, err := p.parse(powLogical)
	if err != nil {
		return nil, err
	}
	a := and{
		left:  left,
		right: right,
	}
	return a, nil
}

func parseOr(p *Parser, left Expr) (Expr, error) {
	p.next()
	right, err := p.parse(powLogical)
	if err != nil {
		return nil, err
	}
	o := or{
		left:  left,
		right: right,
	}
	return o, nil
}
