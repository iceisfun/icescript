package icescript

type TokenKind int

const (
	// Special
	ILLEGAL TokenKind = iota
	EOF
	COMMENT

	// Identifiers + literals
	IDENT  // foo, Bar, _baz
	NUMBER // 123, 3.14
	STRING // "abc", 'x'

	// Keywords
	IF
	ELSE
	FOR
	WHILE
	RETURN
	FUNC
	VAR
	IN
	BREAK
	CONTINUE
	STRUCT
	TYPE
	TRUE
	FALSE
	NULL

	// Operators
	ASSIGN     // =
	PLUS       // +
	MINUS      // -
	STAR       // *
	SLASH      // /
	PERCENT    // %
	BANG       // !
	EQ         // ==
	NEQ        // !=
	LT         // <
	GT         // >
	LTE        // <=
	GTE        // >=
	AND        // &&
	OR         // ||
	PLUSEQ     // +=
	MINUSEQ    // -=
	PLUSPLUS   // ++
	MINUSMINUS // --

	// Delimiters
	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
	LBRACK // [
	RBRACK // ]
	COMMA
	SEMICOLON
	COLON
	DOT
)

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Kind     TokenKind
	Literal  string
	Position Position
}

var keywords = map[string]TokenKind{
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"while":    WHILE,
	"return":   RETURN,
	"func":     FUNC,
	"var":      VAR,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
	"struct":   STRUCT,
	"type":     TYPE,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
}
