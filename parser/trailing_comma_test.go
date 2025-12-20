package parser

import (
	"testing"

	"github.com/iceisfun/icescript/ast"
	"github.com/iceisfun/icescript/lexer"
)

func TestArrayLiteralTrailingComma(t *testing.T) {
	input := `
	[1, 2, 3];
	[1, 2, 3,];
	[1,];
	["one", "two",];
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedLen int
	}{
		{3},
		{3},
		{1},
		{2},
	}

	for i, tt := range tests {
		stmt, ok := program.Statements[i].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[%d] is not ast.ExpressionStatement. got=%T",
				i, program.Statements[i])
		}

		array, ok := stmt.Expression.(*ast.ArrayLiteral)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.ArrayLiteral. got=%T", stmt.Expression)
		}

		if len(array.Elements) != tt.expectedLen {
			t.Errorf("stmt.Expression.Elements has wrong length. got=%d, want=%d",
				len(array.Elements), tt.expectedLen)
		}
	}
}
