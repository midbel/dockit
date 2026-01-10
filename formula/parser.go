package formula

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	powLowest = iota
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

var bindings = map[rune]int{
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

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	prefix map[rune]func() (Expr, error)
	infix  map[rune]func(Expr) (Expr, error)
}

var defaultParser = Parse()

func ParseFormula(str string) (Expr, error) {
	return defaultParser.ParseString(str)
}

func Parse() *Parser {
	var p Parser
	p.prefix = map[rune]func() (Expr, error){
		Ident:   p.parseAdressOrIdentifier,
		Number:  p.parseNumber,
		Literal: p.parseLiteral,
		Sub:     p.parseUnary,
		Add:     p.parseUnary,
		BegGrp:  p.parseGroup,
	}
	p.infix = map[rune]func(Expr) (Expr, error){
		BegGrp: p.parseCall,
		Add:    p.parseBinary,
		Sub:    p.parseBinary,
		Mul:    p.parseBinary,
		Div:    p.parseBinary,
		Concat: p.parseBinary,
		Pow:    p.parseBinary,
		Eq:     p.parseBinary,
		Ne:     p.parseBinary,
		Lt:     p.parseBinary,
		Le:     p.parseBinary,
		Gt:     p.parseBinary,
		Ge:     p.parseBinary,
	}
	return &p
}

func (p *Parser) ParseString(str string) (Expr, error) {
	return p.Parse(strings.NewReader(str))
}

func (p *Parser) Parse(r io.Reader) (Expr, error) {
	scan, err := Scan(r)
	if err != nil {
		return nil, err
	}
	p.scan = scan
	p.next()
	p.next()
	return p.parse(powLowest)
}

func (p *Parser) parse(pow int) (Expr, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unexpected prefix")
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for !p.done() && pow < bindings[p.curr.Type] {
		fn, ok := p.infix[p.curr.Type]
		if !ok {
			return nil, fmt.Errorf("unexpected infix operator")
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseCall(expr Expr) (Expr, error) {
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

func (p *Parser) parseBinary(left Expr) (Expr, error) {
	bin := binary{
		left: left,
		op:   p.curr.Type,
	}
	p.next()
	right, err := p.parse(bindings[bin.op])
	if err != nil {
		return nil, err
	}
	bin.right = right
	return bin, nil
}

func (p *Parser) parseUnary() (Expr, error) {
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

func (p *Parser) parseGroup() (Expr, error) {
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

func (p *Parser) parseNumber() (Expr, error) {
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

func (p *Parser) parseLiteral() (Expr, error) {
	defer p.next()
	i := literal{
		value: p.curr.Literal,
	}
	return i, nil
}

func (p *Parser) parseAdressOrIdentifier() (Expr, error) {
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

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) done() bool {
	return p.curr.Type == Eol
}
