package formula

import (
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
}

type (
	PrefixFunc func(*Parser) (Expr, error)
	InfixFunc  func(*Parser, Expr) (Expr, error)
)

type Grammar struct {
	mode ScanMode

	prefix   map[rune]PrefixFunc
	infix    map[rune]InfixFunc
	bindings map[rune]int

	kwPrefix map[string]PrefixFunc
	kwInfix  map[string]InfixFunc
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
		return nil, fmt.Errorf("unsupported prefix operator")
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
		return nil, fmt.Errorf("unsupported infix operator")
	}
	return fn, nil
}

func (g *Grammar) RegisterInfix(kd rune, fn InfixFunc) {
	g.infix[kd] = fn
}

func (g *Grammar) RegisterInfixKeyword(kw string, fn InfixFunc) {
	g.kwInfix[kw] = fn
}

func (g *Grammar) RegisterPrefix(kd rune, fn PrefixFunc) {
	g.prefix[kd] = fn
}

func (g *Grammar) RegisterPrefixKeyword(kw string, fn PrefixFunc) {
	g.kwPrefix[kw] = fn
}

func (g *Grammar) RegisterBinding(kd rune, pow int) {
	g.bindings[kd] = pow
}

func FormulaGrammar() *Grammar {
	g := Grammar{
		mode:     ModeBasic,
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
	g.mode = ModeScript

	g.RegisterPrefix(BegBlock, parseBlock)

	g.RegisterInfix(BegProp, parseAccess)
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

	return g
}

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	grammar *Grammar
}

func ParseFormula(str string) (Expr, error) {
	p := NewParser(FormulaGrammar())
	return p.ParseString(str)
}

func NewParser(g *Grammar) *Parser {
	var p Parser
	p.grammar = g
	return &p
}

func (p *Parser) ParseString(str string) (Expr, error) {
	return p.Parse(strings.NewReader(str))
}

func (p *Parser) Parse(r io.Reader) (Expr, error) {
	scan, err := Scan(r, p.grammar.mode)
	if err != nil {
		return nil, err
	}
	p.scan = scan
	p.next()
	p.next()
	return p.parse(powLowest)
}

func (p *Parser) ParseNext() (Expr, error) {
	return nil, nil
}

func (p *Parser) parse(pow int) (Expr, error) {
	fn, err := p.grammar.Prefix(p.curr)
	if err != nil {
		return nil, err
	}
	left, err := fn(p)
	if err != nil {
		return nil, err
	}
	for !p.done() && pow < p.grammar.Pow(p.curr.Type) {
		fn, err := p.grammar.Infix(p.curr)
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
	return p.currentLiteral()
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
			return nil, fmt.Errorf("unexpected character in function call")
		}
		c.args = append(c.args, arg)
	}
	if p.curr.Type != EndGrp {
		return nil, fmt.Errorf("unexpected character in function call")
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
	right, err := p.parse(p.grammar.Pow(bin.op))
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
		return nil, fmt.Errorf("missing ')' at end of expression")
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
	defer p.next()
	if p.peek.Type == BegGrp {
		i := identifier{
			name: p.currentLiteral(),
		}
		return i, nil
	}
	var sheet string
	if p.peek.Type == SheetRef {
		sheet = p.currentLiteral()
		p.next()
		p.next()
	}
	a, err := parseCellAddr(p.currentLiteral())
	if err != nil {
		return nil, err
	}
	if p.curr.Type == RangeRef {
		p.next()
	}
	a.Sheet = sheet
	return a, nil
}

func parseBlock(p *Parser) (Expr, error) {
	return nil, nil
}

func parseAccess(p *Parser, left Expr) (Expr, error) {
	return nil, nil
}

func parseAssignment(p *Parser, left Expr) (Expr, error) {
	id, ok := left.(identifier)
	if !ok {
		return nil, fmt.Errorf("identifier expected")
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

func parsePrint(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	_ = expr
	if !p.isEOL() {
		return nil, fmt.Errorf("expected eol")
	}
	p.next()
	return nil, nil
}

func parseSave(p *Parser) (Expr, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	_ = expr
	if !p.isEOL() {
		return nil, fmt.Errorf("expected eol")
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
		return nil, fmt.Errorf("keyword 'to' expected")
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
		return nil, fmt.Errorf("expected eol")
	}
	p.next()
	return nil, nil
}

func parseUse(p *Parser) (Expr, error) {
	return nil, nil
}

func parseImport(p *Parser) (Expr, error) {
	p.next()
	var ref importFile
	switch {
	case p.is(Ident):
		ref.file = identifier{
			name: p.currentLiteral(),
		}
	case p.is(Literal):
		ref.file = literal{
			value: p.currentLiteral(),
		}
	default:
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	if p.is(Keyword) && p.currentLiteral() == kwAs {
		p.next()
		if !p.is(Ident) {
			return nil, fmt.Errorf("unexpected token %s", p.curr)
		}
		ref.alias = identifier{
			name: p.currentLiteral(),
		}
		p.next()
	}
	if !p.isEOL() {
		return nil, fmt.Errorf("expected eol")
	}
	p.next()
	return ref, nil
}
