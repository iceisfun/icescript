package icescript

import "strings"

type Lexer struct {
	input        []rune
	position     int // current position
	readPosition int // next read position
	ch           rune
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		line:  1,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespaceAndComments()

	pos := Position{Line: l.line, Column: l.column}

	switch l.ch {
	case '"', '\'':
		// string literal with escapes
		return l.readString(pos, l.ch)

	// ... keep your existing operator/delimiter cases here ...

	case 0:
		return Token{Kind: EOF, Literal: "", Position: pos}
	}

	// identifiers / keywords
	if isLetter(l.ch) {
		start := l.position
		for isLetter(l.ch) || isDigit(l.ch) {
			l.readChar()
		}
		lit := string(l.input[start:l.position])
		if kw, ok := keywords[lit]; ok {
			return Token{Kind: kw, Literal: lit, Position: pos}
		}
		return Token{Kind: IDENT, Literal: lit, Position: pos}
	}

	// numbers
	if isDigit(l.ch) {
		start := l.position
		dotSeen := false
		for {
			if isDigit(l.ch) {
				l.readChar()
				continue
			}
			if l.ch == '.' && !dotSeen && isDigit(l.peekChar()) {
				dotSeen = true
				l.readChar()
				continue
			}
			break
		}
		return Token{Kind: NUMBER, Literal: string(l.input[start:l.position]), Position: pos}
	}

	// fallback: single-character punct/op as literal
	ch := l.ch
	l.readChar()
	return Token{Kind: ILLEGAL, Literal: string(ch), Position: pos}
}

// Expressions
type ArrayLit struct {
	Elems []Expr
	P     Position
}

func (*ArrayLit) exprNode()       {}
func (a *ArrayLit) Pos() Position { return a.P }

type Field struct {
	Name string
	Expr Expr
	P    Position
}

type ObjectLit struct {
	Fields []Field
	P      Position
}

func (*ObjectLit) exprNode()       {}
func (o *ObjectLit) Pos() Position { return o.P }

type MemberExpr struct {
	Object Expr
	Name   string
	P      Position
}

func (*MemberExpr) exprNode()       {}
func (m *MemberExpr) Pos() Position { return m.P }

type IndexExpr struct {
	Seq   Expr
	Index Expr
	P     Position
}

func (*IndexExpr) exprNode()       {}
func (i *IndexExpr) Pos() Position { return i.P }

// Statements
type VarStmt struct {
	Name string
	Init Expr
	P    Position
}

func (*VarStmt) stmtNode()       {}
func (v *VarStmt) Pos() Position { return v.P }

type AssignStmt struct {
	Name  string // ident only for now
	Value Expr
	P     Position
}

func (*AssignStmt) stmtNode()       {}
func (a *AssignStmt) Pos() Position { return a.P }

type ForInStmt struct {
	VarName  string
	Iterable Expr
	Body     *BlockStmt
	P        Position
}

func (*ForInStmt) stmtNode()       {}
func (f *ForInStmt) Pos() Position { return f.P }

func (l *Lexer) readString(pos Position, quote rune) Token {
	// consume opening quote
	l.readChar()

	var sb strings.Builder
	for {
		switch l.ch {
		case 0, '\n':
			// Unterminated string
			return Token{Kind: ILLEGAL, Literal: string(quote), Position: pos}
		case '\\': // escape
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '\'':
				sb.WriteByte('\'')
			case 'x':
				// \xHH (optional)
				h1, h2 := l.peekChar(), rune(0)
				l.readChar()
				h2 = l.ch
				val := fromHex(h1)<<4 | fromHex(h2)
				if val >= 0 {
					sb.WriteByte(byte(val))
				}
			default:
				// Unknown escape: treat as literal char
				if l.ch != 0 {
					sb.WriteRune(l.ch)
				}
			}
			l.readChar()
		default:
			if l.ch == quote {
				l.readChar() // consume closing quote
				return Token{Kind: STRING, Literal: sb.String(), Position: pos}
			}
			sb.WriteRune(l.ch)
			l.readChar()
		}
	}
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// whitespace
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		}
		// comments
		if l.ch == '/' {
			switch l.peekChar() {
			case '/': // // line comment
				for l.ch != 0 && l.ch != '\n' {
					l.readChar()
				}
				continue
			case '*': // /* block comment */
				l.readChar() // consume '/'
				l.readChar() // consume '*'
				for {
					if l.ch == 0 {
						return
					}
					if l.ch == '*' && l.peekChar() == '/' {
						l.readChar() // '*'
						l.readChar() // '/'
						break
					}
					l.readChar()
				}
				continue
			}
		}
		// neither whitespace nor comment → done
		return
	}
}

func fromHex(r rune) int {
	switch {
	case '0' <= r && r <= '9':
		return int(r - '0')
	case 'a' <= r && r <= 'f':
		return int(r-'a') + 10
	case 'A' <= r && r <= 'F':
		return int(r-'A') + 10
	default:
		return -1
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch rune) bool {
	return ('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z') ||
		ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
