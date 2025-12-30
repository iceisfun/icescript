package lexer

import (
	"testing"

	"github.com/iceisfun/icescript/token"
)

func TestNextToken(t *testing.T) {
	input := `var five = 5;
var ten = 10;
var add = func(x, y) {
	x + y;
};
var result = add(five, ten);
!- / *5;
5 < 10 > 5;
if (5 < 10) {
	return true;
} else {
	return false;
}
10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
3.14
[1:2];
[:1];
[2:];
// single line
/* multi
line */
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.VAR, "var"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.VAR, "var"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "func"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, ";"}, // newline
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.VAR, "var"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, ";"}, // newline
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, ";"}, // newline
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"}, // newline
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.STRING, "foobar"},
		{token.SEMICOLON, ";"}, // newline
		{token.STRING, "foo bar"},
		{token.SEMICOLON, ";"}, // newline
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"}, // newline
		{token.FLOAT, "3.14"},
		{token.SEMICOLON, ";"}, // newline
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COLON, ":"},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.LBRACKET, "["},
		{token.COLON, ":"},
		{token.INT, "1"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.LBRACKET, "["},
		{token.INT, "2"},
		{token.COLON, ":"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.SEMICOLON, ";"}, // newline
		{token.SEMICOLON, ";"}, // single line comment newline
		{token.SEMICOLON, ";"}, // multi line comment newline
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStringEscapeSequences(t *testing.T) {
	input := `
"newline\n"
"tab\t"
"quote\""
"backslash\\"
"mixed\n\t\"\\"
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.SEMICOLON, ";"}, // initial newline
		{token.STRING, "newline\n"},
		{token.SEMICOLON, ";"},
		{token.STRING, "tab\t"},
		{token.SEMICOLON, ";"},
		{token.STRING, "quote\""},
		{token.SEMICOLON, ";"},
		{token.STRING, "backslash\\"},
		{token.SEMICOLON, ";"},
		{token.STRING, "mixed\n\t\"\\"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
