package formula

import (
	"bytes"
	"io"
	"unicode/utf8"
)

const (
	Invalid rune = -(1 << iota)
	Eol
	Ident
	Number
	Literal
	Add
	Sub
	Mul
	Div
	Percent
	Pow
	Concat
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	Comma
	BegGrp
	EndGrp
	BegArr
	EndArr
	RangeRef
	SheetRef
)

type Token struct {
	Literal  string
	Type     rune
	Position int
}

type Scanner struct {
	input []byte
	pos   int
	next  int
	char  rune

	buf bytes.Buffer
}

func Scan(r io.Reader) (*Scanner, error) {
	var (
		scan Scanner
		err  error
	)
	scan.input, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	scan.read()
	if scan.char == equal {
		scan.read()
	}
	// scan.read()
	return &scan, nil
}

func (s *Scanner) Scan() Token {
	s.skipBlanks()

	var tok Token
	if s.done() {
		tok.Type = Eol
		return tok
	}
	defer s.reset()
	switch {
	case isOperator(s.char):
		s.scanOperator(&tok)
	case isDelimiter(s.char):
		s.scanDelimiter(&tok)
	case isQuote(s.char):
		s.scanLiteral(&tok)
	case isDigit(s.char):
		s.scanNumber(&tok)
	default:
		s.scanIdent(&tok)
	}
	return tok
}

func (s *Scanner) scanIdent(tok *Token) {
	for !s.done() && isAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Type = Ident
	tok.Literal = s.literal()
}

func (s *Scanner) scanNumber(tok *Token) {
	tok.Type = Number
	for !s.done() && isDigit(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
	if s.char != dot {
		return
	}
	s.write()
	s.read()
	for !s.done() && isDigit(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
}

func (s *Scanner) scanLiteral(tok *Token) {
	quote := s.char
	s.read()
	for !s.done() && !isQuote(s.char) {
		s.write()
		s.read()
	}
	tok.Type = Literal
	tok.Literal = s.literal()
	if isQuote(s.char) && quote == s.char {
		s.read()
	} else {
		tok.Type = Invalid
	}
}

func (s *Scanner) scanOperator(tok *Token) {
	tok.Type = Invalid
	switch s.char {
	case amper:
		tok.Type = Concat
	case percent:
		tok.Type = Percent
	case plus:
		tok.Type = Add
	case minus:
		tok.Type = Sub
	case star:
		tok.Type = Mul
	case slash:
		tok.Type = Div
	case caret:
		tok.Type = Pow
	case langle:
		tok.Type = Lt
		s.read()
		if s.char == equal {
			tok.Type = Le
		} else if s.char == rangle {
			tok.Type = Ne
		}
	case rangle:
		tok.Type = Gt
		s.read()
		if s.char == equal {
			tok.Type = Ge
		}
	case equal:
		tok.Type = Eq
	case colon:
		tok.Type = RangeRef
	case bang:
		tok.Type = SheetRef
	default:
	}
	s.read()
}

func (s *Scanner) scanDelimiter(tok *Token) {
	tok.Type = Invalid
	switch s.char {
	case semi, comma:
		tok.Type = Comma
	case lparen:
		tok.Type = BegGrp
	case rparen:
		tok.Type = EndGrp
	case lcurly:
		tok.Type = BegArr
	case rcurly:
		tok.Type = EndArr
	default:
	}
	s.read()
}

func (s *Scanner) literal() string {
	return s.buf.String()
}

func (s *Scanner) write() {
	s.buf.WriteRune(s.char)
}

func (s *Scanner) reset() {
	s.buf.Reset()
}

func (s *Scanner) read() {
	if s.pos >= len(s.input) {
		s.char = 0
		return
	}
	r, n := utf8.DecodeRune(s.input[s.next:])
	if r == utf8.RuneError {
		s.char = 0
		s.next = len(s.input)
	}
	s.char, s.pos, s.next = r, s.next, s.next+n
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func (s *Scanner) done() bool {
	return s.pos >= len(s.input) && s.char == 0
}

func (s *Scanner) skipBlanks() {
	for isBlank(s.char) {
		s.read()
	}
}

const (
	underscore = '_'
	bang       = '!'
	semi       = ';'
	comma      = ','
	rparen     = ')'
	lparen     = '('
	lcurly     = '{'
	rcurly     = '}'
	squote     = '\''
	dquote     = '"'
	space      = ' '
	tab        = '\t'
	plus       = '+'
	minus      = '-'
	star       = '*'
	slash      = '/'
	caret      = '^'
	equal      = '='
	langle     = '<'
	rangle     = '>'
	colon      = ':'
	dot        = '.'
	amper      = '&'
	percent    = '%'
	dollar     = '$'
)

func isQuote(c rune) bool {
	return c == squote || c == dquote
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLetter(c rune) bool {
	return isLower(c) || isUpper(c) || c == underscore
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return isLetter(c) || isDigit(c) || c == dollar
}

func isBlank(c rune) bool {
	return c == space || c == tab
}

func isDelimiter(c rune) bool {
	return c == semi || c == lparen || c == rparen ||
		c == lcurly || c == rcurly || c == comma
}

func isOperator(c rune) bool {
	return c == plus || c == minus || c == slash || c == star ||
		c == langle || c == rangle || c == colon || c == bang ||
		c == equal || c == caret || c == amper || c == percent
}
