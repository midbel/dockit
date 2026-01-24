package formula

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	Invalid rune = 0

	EOF rune = 1 << iota
	Eol
	Keyword
	Ident
	Number
	Literal
	Comment
	Assign
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
	Dot
	Begin
	End
	RangeRef
	SheetRef
)

const (
	groupTok rune = 1 << iota
	blockTok
	propTok
)

const (
	BegBlock = blockTok | Begin
	EndBlock = blockTok | End
	BegGrp   = groupTok | Begin
	EndGrp   = groupTok | End
	BegProp  = propTok | Begin
	EndProp  = propTok | End
)

const (
	AddAssign    = Add | Assign
	DivAssign    = Div | Assign
	SubAssign    = Sub | Assign
	MulAssign    = Mul | Assign
	PowAssign    = Pow | Assign
	ConcatAssign = Concat | Assign
)

const (
	kwImport  = "import"
	kwUse     = "use"
	kwWith    = "with"
	kwPrint   = "print"
	kwSave    = "save"
	kwExport  = "export"
	kwDefault = "default"
	kwFrom    = "from"
	kwIn      = "in"
	kwAs      = "as"
	kwTo      = "to"
	kwEnd     = "end"
)

const (
	kwSheet  = "sheet"
	kwChart  = "chart"
	kwPivot  = "pivot"
	kwFilter = "filter"
)

func isKeyword(str string) bool {
	switch str {
	case kwImport:
	case kwUse:
	case kwWith:
	case kwFrom:
	case kwPrint:
	case kwSave:
	case kwExport:
	case kwDefault:
	case kwIn:
	case kwAs:
	case kwTo:
	case kwEnd:
	case kwSheet:
	case kwChart:
	case kwPivot:
	case kwFilter:
	default:
		return false
	}
	return true
}

type ScanMode int8

const (
	ModeFormula ScanMode = 1 << iota
	ModeScript
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Literal string
	Type    rune
	Position
}

func (t Token) String() string {
	var str string
	switch t.Type {
	case Invalid:
		return "<invalid>"
	case EOF:
		return "<eof>"
	case Eol:
		return "<eol>"
	case Keyword:
		str = "keyword"
	case Ident:
		str = "identifier"
	case Number:
		str = "number"
	case Literal:
		str = "literal"
	case Comment:
		str = "comment"
	case Assign:
		return "<assignment>"
	case Add:
		return "<add>"
	case Sub:
		return "<subtract>"
	case Mul:
		return "<multiply>"
	case Div:
		return "<divide>"
	case Percent:
		return "<percent>"
	case Pow:
		return "<power>"
	case Concat:
		return "<concat>"
	case Eq:
		return "<equal>"
	case Ne:
		return "<notequal>"
	case Lt:
		return "<lesser>"
	case Le:
		return "<lesseq>"
	case Gt:
		return "<greater>"
	case Ge:
		return "<greateq>"
	case Comma:
		return "<comma>"
	case Dot:
		return "<dot>"
	case BegGrp:
		return "<beg-group>"
	case EndGrp:
		return "<end-group>"
	case BegProp:
		return "<beg-prop>"
	case EndProp:
		return "<end-prop>"
	case BegBlock:
		return "<beg-block>"
	case EndBlock:
		return "<end-block>"
	case RangeRef:
		return "<range>"
	case SheetRef:
		return "<sheet>"
	case AddAssign:
		return "<add-assign>"
	case DivAssign:
		return "<div-assign>"
	case SubAssign:
		return "<sub-assign>"
	case MulAssign:
		return "<mul-assign>"
	case PowAssign:
		return "<pow-assign>"
	case ConcatAssign:
		return "<concat-assign>"
	}
	return fmt.Sprintf("%s(%s)", str, t.Literal)
}

type Scanner struct {
	input []byte
	pos   int
	next  int
	char  rune

	Position

	buf  bytes.Buffer
	mode ScanMode
}

func Scan(r io.Reader, mode ScanMode) (*Scanner, error) {
	var (
		scan Scanner
		err  error
	)
	scan.mode = mode
	scan.input, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	scan.Position.Line = 1
	scan.read()
	if scan.char == equal {
		scan.read()
	}
	return &scan, nil
}

func (s *Scanner) Scan() Token {
	s.skipBlanks()

	var tok Token
	tok.Position = s.Position
	if s.done() {
		tok.Type = EOF
		return tok
	}
	defer s.reset()
	switch {
	case isNL(s.char) && s.mode == ModeScript:
		s.scanNL(&tok)
	case isComment(s.char) && s.mode == ModeScript:
		s.scanComment(&tok)
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

func (s *Scanner) scanNL(tok *Token) {
	s.skipNL()
	tok.Type = Eol
}

func (s *Scanner) scanComment(tok *Token) {
	s.read()
	s.skipBlanks()
	for !s.done() && !isNL(s.char) {
		s.write()
		s.read()
	}
	s.skipNL()
	tok.Type = Comment
	tok.Literal = s.literal()
}

func (s *Scanner) scanIdent(tok *Token) {
	for !s.done() && isAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Type = Ident
	tok.Literal = s.literal()
	if s.allowKeywords() && isKeyword(tok.Literal) {
		tok.Type = Keyword
	}
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
	case dot:
		tok.Type = Dot
	case amper:
		tok.Type = Concat
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = ConcatAssign
		}
	case percent:
		tok.Type = Percent
	case plus:
		tok.Type = Add
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = AddAssign
		}
	case minus:
		tok.Type = Sub
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = SubAssign
		}
	case star:
		tok.Type = Mul
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = MulAssign
		}
	case slash:
		tok.Type = Div
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = DivAssign
		}
	case caret:
		tok.Type = Pow
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = PowAssign
		}
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
		if s.peek() == equal && s.mode == ModeScript {
			s.read()
			tok.Type = Assign
		}
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
		tok.Type = BegBlock
	case rcurly:
		tok.Type = EndBlock
	case lsquare:
		tok.Type = BegProp
	case rsquare:
		tok.Type = EndProp
	default:
	}
	s.read()
}

func (s *Scanner) allowKeywords() bool {
	return s.mode == ModeScript
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

	if s.char == nl {
		s.Line += 1
		s.Column = 0
	}
	s.Column++
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func (s *Scanner) done() bool {
	return s.pos >= len(s.input) || s.char == 0
}

func (s *Scanner) skipNL() {
	for isNL(s.char) {
		s.read()
	}
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
	nl         = '\n'
	cr         = '\r'
	pound      = '#'
	lsquare    = '['
	rsquare    = ']'
)

func isComment(c rune) bool {
	return c == pound
}

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

func isNL(c rune) bool {
	return c == nl || c == cr
}

func isDelimiter(c rune) bool {
	return c == semi || c == lparen || c == rparen ||
		c == lcurly || c == rcurly || c == comma ||
		c == lsquare || c == rsquare
}

func isOperator(c rune) bool {
	return c == plus || c == minus || c == slash || c == star ||
		c == langle || c == rangle || c == colon || c == bang ||
		c == equal || c == caret || c == amper || c == percent ||
		c == dot
}
