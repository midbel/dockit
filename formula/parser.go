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
	powCall
)

var defaultBindings = map[rune]int{
	AddAssign: powAssign,
	SubAssign: powAssign,
	MulAssign: powAssign,
	DivAssign: powAssign,
	PowAssign: powAssign,
	ConcatAssign: powAssign,
	Add:     powAdd,
	Sub:     powAdd,
	Mul:     powMul,
	Div:     powMul,
	Percent: powPercent,
	Pow:     powPow,
	Concat:  powConcat,
	Eq:      powEq,
	Ne:      powEq,
	Lt:      powCmp,
	Le:      powCmp,
	Gt:      powCmp,
	Ge:      powCmp,
	BegGrp:  powCall,
}

type (
	PrefixFunc func(*Parser) (Expr, error)
	InfixFunc  func(*Parser, Expr) (Expr, error)
)

type Grammar struct {
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

	g.RegisterPrefix(BegBlock, parseBlock)
	g.RegisterPrefix(Keyword, parseKeyword)

	g.RegisterPrefixKeyword(kwLet, parseLet)
	g.RegisterPrefixKeyword(kwPrint, parsePrint)
	g.RegisterPrefixKeyword(kwImport, parseImport)

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
	scan, err := Scan(r, ModeBasic)
	if err != nil {
		return nil, err
	}
	p.scan = scan
	p.next()
	p.next()
	return p.parse(powLowest)
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
	return p.curr.Type == EOF
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

	x, err := strconv.ParseFloat(p.curr.Literal, 64)
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
		value: p.curr.Literal,
	}
	return i, nil
}

func parseAdressOrIdentifier(p *Parser) (Expr, error) {
	defer p.next()
	if p.peek.Type == BegGrp {
		i := identifier{
			name: p.curr.Literal,
		}
		return i, nil
	}
	var sheet string
	if p.peek.Type == SheetRef {
		sheet = p.curr.Literal
		p.next()
		p.next()
	}
	a, err := parseCellAddr(p.curr.Literal)
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

func parseKeyword(p *Parser) (Expr, error) {
	return nil, nil
}

func parseAssign(p *Parser, left Expr) (Expr, error) {
	return nil, nil
}

func parseLet(p *Parser) (Expr, error) {
	return nil, nil
}

func parsePrint(p *Parser) (Expr, error) {
	return nil, nil
}

func parseImport(p *Parser) (Expr, error) {
	return nil, nil
}
