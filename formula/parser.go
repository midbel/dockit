package formula

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"
)

const (
	powLowest = iota
	powAssign
	powEq
	powCmp
	powConcat
	powAdd
	powMul
	powPow
	powUnary
	powPercent
	powProp
	powCall
)

var defaultBindings = map[rune]int{
	AddAssign:    powAssign,
	SubAssign:    powAssign,
	MulAssign:    powAssign,
	DivAssign:    powAssign,
	PowAssign:    powAssign,
	ConcatAssign: powAssign,
	Assign:       powAssign,
	Add:          powAdd,
	Sub:          powAdd,
	Mul:          powMul,
	Div:          powMul,
	Percent:      powPercent,
	Pow:          powPow,
	Concat:       powConcat,
	Eq:           powEq,
	Ne:           powEq,
	Lt:           powCmp,
	Le:           powCmp,
	Gt:           powCmp,
	Ge:           powCmp,
	BegGrp:       powCall,
	BegProp:      powProp,
	Dot:          powProp,
}

type (
	PrefixFunc func(*Parser) (Expr, error)
	InfixFunc  func(*Parser, Expr) (Expr, error)
)

var errForbidden = fmt.Errorf("not allowed")

func forbiddenInfix(_ *Parser, _ Expr) (Expr, error) {
	return nil, fmt.Errorf("%w: infix operator/keyword", errForbidden)
}

func forbiddenPrefix(_ *Parser, _ Expr) (Expr, error) {
	return nil, fmt.Errorf("%w: prefix operator/keyword", errForbidden)
}

type Grammar struct {
	name string
	mode ScanMode

	prefix   map[rune]PrefixFunc
	infix    map[rune]InfixFunc
	bindings map[rune]int

	kwPrefix map[string]PrefixFunc
	kwInfix  map[string]InfixFunc
}

func (g *Grammar) Context() string {
	return g.name
}

func (g *Grammar) Pow(kind rune) int {
	pow, ok := g.bindings[kind]
	if !ok {
		pow = powLowest
	}
	return pow
}

func (g *Grammar) Prefix(tok Token) (PrefixFunc, error) {
	if tok.Type == Keyword {
		fn, ok := g.kwPrefix[tok.Literal]
		if !ok {
			return nil, fmt.Errorf("%s: unknown keyword", tok.Literal)
		}
		return fn, nil
	}
	fn, ok := g.prefix[tok.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported prefix operator (%s)", tok)
	}
	return fn, nil
}

func (g *Grammar) Infix(tok Token) (InfixFunc, error) {
	if tok.Type == Keyword {
		fn, ok := g.kwInfix[tok.Literal]
		if !ok {
			return nil, fmt.Errorf("%s: unknown keyword", tok.Literal)
		}
		return fn, nil
	}
	fn, ok := g.infix[tok.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported infix operator (%s)", tok)
	}
	return fn, nil
}

func (g *Grammar) RegisterInfix(kd rune, fn InfixFunc) {
	g.infix[kd] = fn
}

func (g *Grammar) UnregisterInfix(kd rune) {
	g.infix[kd] = fobiddenInfix
}

func (g *Grammar) RegisterInfixKeyword(kw string, fn InfixFunc) {
	g.kwInfix[kw] = fn
}

func (g *Grammar) UnregisterInfixKeyword(kw string) {
	g.kwInfix[kw] = fobiddenInfix
}

func (g *Grammar) RegisterPrefix(kd rune, fn PrefixFunc) {
	g.prefix[kd] = fn
}

func (g *Grammar) UnregisterPrefix(kd rune) {
	g.prefix[kd] = forbiddenPrefix
}

func (g *Grammar) RegisterPrefixKeyword(kw string, fn PrefixFunc) {
	g.kwPrefix[kw] = fn
}

func (g *Grammar) UnregisterPrefixKeyword(kw string) {
	g.kwPrefix[kw] = fobiddenPrefix
}

func (g *Grammar) RegisterBinding(kd rune, pow int) {
	g.bindings[kd] = pow
}

func FormulaGrammar() *Grammar {
	g := Grammar{
		name:     "formula",
		mode:     ModeFormula,
		prefix:   make(map[rune]PrefixFunc),
		kwPrefix: make(map[string]PrefixFunc),
		infix:    make(map[rune]InfixFunc),
		kwInfix:  make(map[string]InfixFunc),
		bindings: maps.Clone(defaultBindings),
	}
	g.RegisterPrefix(Ident, parseAdressOrIdentifier)
	g.RegisterPrefix(Number, parseNumber)
	g.RegisterPrefix(Literal, parseLiteral)
	g.RegisterPrefix(Sub, parseUnary)
	g.RegisterPrefix(Add, parseUnary)
	g.RegisterPrefix(BegGrp, parseGroup)

	g.RegisterInfix(BegGrp, parseCall)
	g.RegisterInfix(Add, parseBinary)
	g.RegisterInfix(Sub, parseBinary)
	g.RegisterInfix(Mul, parseBinary)
	g.RegisterInfix(Div, parseBinary)
	g.RegisterInfix(Concat, parseBinary)
	g.RegisterInfix(Pow, parseBinary)
	g.RegisterInfix(Eq, parseBinary)
	g.RegisterInfix(Ne, parseBinary)
	g.RegisterInfix(Lt, parseBinary)
	g.RegisterInfix(Le, parseBinary)
	g.RegisterInfix(Gt, parseBinary)
	g.RegisterInfix(Ge, parseBinary)

	return &g
}

func ScriptGrammar() *Grammar {
	g := FormulaGrammar()
	g.name = "script"
	g.mode = ModeScript

	g.RegisterPrefix(BegBlock, parseBlock)
	g.RegisterPrefix(Eq, parseLambda)

	// g.RegisterInfix(BegProp, parseIndex)
	g.RegisterInfix(Dot, parseAccess)
	g.RegisterInfix(Assign, parseAssignment)
	g.RegisterInfix(AddAssign, parseAssignment)
	g.RegisterInfix(SubAssign, parseAssignment)
	g.RegisterInfix(MulAssign, parseAssignment)
	g.RegisterInfix(PowAssign, parseAssignment)
	g.RegisterInfix(DivAssign, parseAssignment)

	g.RegisterPrefixKeyword(kwUse, parseUse)
	g.RegisterPrefixKeyword(kwImport, parseImport)
	g.RegisterPrefixKeyword(kwPrint, parsePrint)
	g.RegisterPrefixKeyword(kwSave, parseSave)
	g.RegisterPrefixKeyword(kwExport, parseExport)
	g.RegisterPrefixKeyword(kwWith, parseWith)
	g.RegisterPrefixKeyword(kwDefault, parseDefault)

	return g
}

type GrammarStack []*Grammar

func (gs *GrammarStack) Top() *Grammar {
	n := len(*gs)
	return (*gs)[n-1]
}

func (gs *GrammarStack) Mode() ScanMode {
	return gs.Top().mode
}

func (gs *GrammarStack) Context() string {
	return gs.Top().Context()
}

func (gs *GrammarStack) Pow(kind rune) int {
	for i := len(*gs) - 1; i >= 0; i-- {
		pow := (*gs)[i].Pow(kind)
		if pow > powLowest {
			return pow
		}
	}
	return powLowest
}

func (gs *GrammarStack) Prefix(tok Token) (PrefixFunc, error) {
	var lastErr error
	for i := len(*gs) - 1; i >= 0; i-- {
		fn, err := (*gs)[i].Prefix(tok)
		if err == nil {
			return fn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (gs *GrammarStack) Infix(tok Token) (InfixFunc, error) {
	var lastErr error
	for i := len(*gs) - 1; i >= 0; i-- {
		fn, err := (*gs)[i].Infix(tok)
		if err == nil {
			return fn, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (gs *GrammarStack) Pop() {
	n := len(*gs)
	if n > 1 {
		*gs = (*gs)[:n-1]
	}
}

func (gs *GrammarStack) Push(g *Grammar) {
	*gs = append(*gs, g)
}

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	stack *GrammarStack
}

func ParseFormula(str string) (Expr, error) {
	p := NewParser(FormulaGrammar())
	return p.ParseString(str)
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
	if p.is(Comment) {
		p.next()
		return p.ParseNext()
	}
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
	var list []Expr
	for !p.done() {
		if p.is(Comment) {
			p.next()
			continue
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	_ = list
	return nil, nil
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
	return p.is(EOF)
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) isEOL() bool {
	return p.is(Eol)
}

func (p *Parser) currentLiteral() string {
	return p.curr.Literal
}

func (p *Parser) pow(kind rune) int {
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
	return fmt.Errorf("%s: %s", p.currGrammar().Context(), msg)
}

func parseCall(p *Parser, expr Expr) (Expr, error) {
	p.next()
	c := call{
		ident: expr,
	}
	for !p.done() && p.curr.Type != EndGrp {
		arg, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		switch p.curr.Type {
		case Comma:
			p.next()
		case EndGrp:
		default:
			return nil, p.makeError("unexpected character in function call")
		}
		c.args = append(c.args, arg)
	}
	if p.curr.Type != EndGrp {
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
	if p.curr.Type != EndGrp {
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
	if p.peek.Type == BegGrp || p.peek.Type == Dot {
		id := identifier{
			name: p.currentLiteral(),
		}
		p.next()
		return id, nil
	}

	var sheet string
	if p.peek.Type == SheetRef {
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

	if p.is(RangeRef) {
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
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	a := access{
		expr:   left,
		member: expr,
	}
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
	op := p.curr.Type
	p.next()

	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	switch op {
	case Assign:
	case AddAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    Add,
		}
		expr = b
	case SubAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    Sub,
		}
		expr = b
	case MulAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    Mul,
		}
		expr = b
	case DivAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    Div,
		}
		expr = b
	case PowAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    Pow,
		}
		expr = b
	case ConcatAssign:
		b := binary{
			left:  left,
			right: expr,
			op:    Concat,
		}
		expr = b
	default:
	}
	a.expr = expr
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
	_ = expr
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
	return nil, nil
}

func parseExport(p *Parser) (Expr, error) {
	p.next()
	ident, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	_ = ident
	if !p.is(Keyword) && p.currentLiteral() != kwTo {
		return nil, p.makeError("keyword 'to' expected")
	}
	p.next()
	file, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	_ = file
	if p.is(Keyword) && p.currentLiteral() == kwAs {
		p.next()
		format, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		_ = format
	}
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	p.next()
	return nil, nil
}

func parseUse(p *Parser) (Expr, error) {
	p.next()
	var stmt useFile
	switch {
	case p.is(Ident):
		stmt.file = identifier{
			name: p.currentLiteral(),
		}
	case p.is(Literal):
		stmt.file = literal{
			value: p.currentLiteral(),
		}
	default:
		msg := fmt.Sprintf("unexpected token %s", p.curr)
		return nil, p.makeError(msg)
	}
	p.next()
	if p.is(Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(Ident) {
			msg := fmt.Sprintf("unexpected token %s", p.curr)
			return nil, p.makeError(msg)
		}
		stmt.alias = identifier{
			name: p.currentLiteral(),
		}
		p.next()
	}
	if !p.isEOL() {
		return nil, p.makeError("expected eol at end of use")
	}
	p.next()
	return stmt, nil
}

func parseImport(p *Parser) (Expr, error) {
	p.next()
	var stmt importFile
	switch {
	case p.is(Ident):
		stmt.file = identifier{
			name: p.currentLiteral(),
		}
	case p.is(Literal):
		stmt.file = literal{
			value: p.currentLiteral(),
		}
	default:
		msg := fmt.Sprintf("unexpected token %s", p.curr)
		return nil, p.makeError(msg)
	}
	p.next()
	if p.is(Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(Ident) {
			msg := fmt.Sprintf("unexpected token %s", p.curr)
			return nil, p.makeError(msg)
		}
		stmt.alias = identifier{
			name: p.currentLiteral(),
		}
		p.next()
	}
	if !p.isEOL() {
		return nil, p.makeError("expected eol at end of import")
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
	if !p.is(Keyword) {
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
		isEnd := p.is(Keyword) && p.currentLiteral() == kwEnd
		return !isEnd
	})
	if err != nil {
		return nil, err
	}
	if !p.is(Keyword) && p.currentLiteral() != kwEnd {
		return nil, p.makeError("expected 'end' keyword")
	}
	p.next()
	if !p.isEOL() {
		return nil, p.makeError("expected eol")
	}
	return get(body), nil
}
