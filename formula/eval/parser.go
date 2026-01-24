package eval

import (
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/value"
)

func FormulaGrammar() *Grammar {
	g := Grammar{
		name:     "formula",
		mode:     ModeFormula,
		prefix:   make(map[op.Op]PrefixFunc),
		kwPrefix: make(map[string]PrefixFunc),
		infix:    make(map[op.Op]InfixFunc),
		kwInfix:  make(map[string]InfixFunc),
		bindings: maps.Clone(defaultBindings),
	}
	g.RegisterPrefix(op.Ident, parseAdressOrIdentifier)
	g.RegisterPrefix(op.Number, parseNumber)
	g.RegisterPrefix(op.Literal, parseLiteral)
	g.RegisterPrefix(op.Sub, parseUnary)
	g.RegisterPrefix(op.Add, parseUnary)
	g.RegisterPrefix(op.BegGrp, parseGroup)

	g.RegisterInfix(op.BegGrp, parseCall)
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

	return &g
}

func ScriptGrammar() *Grammar {
	g := FormulaGrammar()
	g.name = "script"
	g.mode = ModeScript

	g.RegisterPrefix(op.BegBlock, parseBlock)
	g.RegisterPrefix(op.Eq, parseLambda)

	// g.RegisterInfix(BegProp, parseIndex)
	g.RegisterInfix(op.Dot, parseAccess)
	g.RegisterInfix(op.Assign, parseAssignment)
	g.RegisterInfix(op.AddAssign, parseAssignment)
	g.RegisterInfix(op.SubAssign, parseAssignment)
	g.RegisterInfix(op.MulAssign, parseAssignment)
	g.RegisterInfix(op.PowAssign, parseAssignment)
	g.RegisterInfix(op.DivAssign, parseAssignment)

	g.RegisterPrefixKeyword(kwUse, parseUse)
	g.RegisterPrefixKeyword(kwImport, parseImport)
	g.RegisterPrefixKeyword(kwPrint, parsePrint)
	g.RegisterPrefixKeyword(kwSave, parseSave)
	g.RegisterPrefixKeyword(kwExport, parseExport)
	g.RegisterPrefixKeyword(kwWith, parseWith)
	g.RegisterPrefixKeyword(kwDefault, parseDefault)

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
	f := formula{
		Expr: expr,
	}
	return f, nil
}

func NewParser(g *Grammar) *Parser {
	var p Parser
	p.stack = new(GrammarStack)
	p.pushGrammar(g)
	return &p
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
		p.skipEOL()
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
	for !p.done() && pow < p.pow(p.curr.Type) {
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

func (p *Parser) isEOL() bool {
	return p.is(op.Eol) || p.is(op.EOF)
}

func (p *Parser) skipEOL() {
	for p.isEOL() {
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

func (p *Parser) infix() (InfixFunc, error) {
	return p.currGrammar().Infix(p.curr)
}

func (p *Parser) pushGrammar(g *Grammar) {
	p.stack.Push(g)
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

func parseCall(p *Parser, expr Expr) (Expr, error) {
	p.next()
	c := call{
		ident: expr,
	}
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
		c.args = append(c.args, arg)
	}
	if !p.is(op.EndGrp) {
		return nil, p.makeError("unexpected character in function call")
	}
	p.next()
	return c, nil
}

func parseBinary(p *Parser, left Expr) (Expr, error) {
	bin := binary{
		left: left,
		op:   p.curr.Type,
	}
	p.next()
	right, err := p.parse(p.pow(bin.op))
	if err != nil {
		return nil, err
	}
	bin.right = right
	return bin, nil
}

func parseUnary(p *Parser) (Expr, error) {
	una := unary{
		op: p.curr.Type,
	}
	p.next()
	right, err := p.parse(powUnary)
	if err != nil {
		return nil, err
	}
	una.right = right
	return una, nil
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
	n := number{
		value: x,
	}
	return n, nil
}

func parseLiteral(p *Parser) (Expr, error) {
	defer p.next()
	i := literal{
		value: p.currentLiteral(),
	}
	return i, nil
}

func parseAdressOrIdentifier(p *Parser) (Expr, error) {
	if p.peek.Type == op.BegGrp || p.peek.Type == op.Dot {
		id := identifier{
			name: p.currentLiteral(),
		}
		p.next()
		return id, nil
	}

	var sheet string
	if p.peek.Type == op.SheetRef {
		sheet = p.currentLiteral()
		p.next()
		p.next()
	}

	start, err := parseCellAddr(p.currentLiteral())
	if err != nil {
		id := identifier{
			name: p.currentLiteral(),
		}
		p.next()
		return id, nil
	}

	start.Sheet = sheet
	p.next()

	if p.is(op.RangeRef) {
		p.next()

		end, err := parseCellAddr(p.currentLiteral())
		if err != nil {
			return nil, err
		}
		end.Sheet = sheet
		p.next()

		rg := rangeAddr{
			startAddr: start,
			endAddr:   end,
		}
		return rg, nil
	}
	return start, nil
}

func parseBlock(p *Parser) (Expr, error) {
	return nil, nil
}

func parseLambda(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	e := lambda{
		expr: expr,
	}
	return e, nil
}

// func parseIndex(p *Parser, left Expr) (Expr, error) {
// 	return nil, nil
// }

func parseAccess(p *Parser, left Expr) (Expr, error) {
	p.next()
	if !p.is(op.Ident) {
		return nil, p.makeError("identifier expected")
	}
	a := access{
		expr: left,
		prop: p.currentLiteral(),
	}
	p.next()
	return a, nil
}

func parseAssignment(p *Parser, left Expr) (Expr, error) {
	id, ok := left.(identifier)
	if !ok {
		return nil, p.makeError("identifier expected")
	}
	a := assignment{
		ident: id,
	}
	oper := p.curr.Type
	p.next()

	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	switch oper {
	case op.Assign:
	case op.AddAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    op.Add,
		}
		expr = b
	case op.SubAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    op.Sub,
		}
		expr = b
	case op.MulAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    op.Mul,
		}
		expr = b
	case op.DivAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    op.Div,
		}
		expr = b
	case op.PowAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    op.Pow,
		}
		expr = b
	case op.ConcatAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    op.Concat,
		}
		expr = b
	default:
	}
	a.expr = expr
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
	return a, nil
}

func parseDefault(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	stmt := defaultRef{
		expr: expr,
	}
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
	return stmt, nil
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
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
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
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
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
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
	return stmt, nil
}

func parseUse(p *Parser) (Expr, error) {
	p.next()
	var stmt useFile
	if !p.is(op.Literal) {
		msg := fmt.Sprintf("literal expected instead of %s", p.curr)
		return nil, p.makeError(msg)
	}
	stmt.file = p.currentLiteral()
	p.next()
	if p.is(op.Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(op.Ident) {
			msg := fmt.Sprintf("literal/identifier expected instead of %s", p.curr)
			return nil, p.makeError(msg)
		}
		stmt.alias = p.currentLiteral()
		p.next()
	}
	if !p.isEOL() {
		return nil, p.makeError("eol expected")
	}
	p.next()
	return stmt, nil
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
	if p.is(op.Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(op.Ident) {
			msg := fmt.Sprintf("literal/identifier expected instead of %s", p.curr)
			return nil, p.makeError(msg)
		}
		stmt.alias = p.currentLiteral()
		p.next()
	}
	if !p.isEOL() {
		return nil, p.makeError("eol expected")
	}
	p.next()
	return stmt, nil
}

func chartGrammar() *Grammar {
	return nil
}

func pivotGrammar() *Grammar {
	return nil
}

func sheetGrammar() *Grammar {
	return nil
}

func filterGrammar() *Grammar {
	return nil
}

func parseWith(p *Parser) (Expr, error) {
	p.next()
	if !p.is(op.Keyword) {
		return nil, p.makeError("keyword expected")
	}
	var get func([]Expr) Expr
	switch p.currentLiteral() {
	case kwSheet:
		p.pushGrammar(sheetGrammar())
		get = makeSheetExpr
	case kwChart:
		p.pushGrammar(chartGrammar())
		get = makeChartExpr
	case kwPivot:
		p.pushGrammar(pivotGrammar())
		get = makePivotExpr
	case kwFilter:
		p.pushGrammar(filterGrammar())
		get = makeFilterExpr
	default:
		msg := fmt.Sprintf("unexpected keyword: %s", p.currentLiteral())
		return nil, p.makeError(msg)
	}
	defer p.popGrammar()
	p.next()

	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()

	body, err := p.parseUntil(func() bool {
		isEnd := p.is(op.Keyword) && p.currentLiteral() == kwEnd
		return !isEnd
	})
	if err != nil {
		return nil, err
	}
	if !p.is(op.Keyword) && p.currentLiteral() != kwEnd {
		return nil, p.makeError("expected 'end' keyword")
	}
	p.next()
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	return get(body), nil
}
