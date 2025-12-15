package compiler

import (
	"fmt"
	"testing"

	"github.com/iceisfun/icescript/ast"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/opcode"
	"github.com/iceisfun/icescript/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []any
	expectedInstructions []code
}

type code struct {
	op       opcode.Opcode
	operands []int
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code{
				{opcode.OpConstant, []int{0}},
				{opcode.OpConstant, []int{1}},
				{opcode.OpAdd, []int{}},
				{opcode.OpPop, []int{}},
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code{
				{opcode.OpConstant, []int{0}},
				{opcode.OpPop, []int{}},
				{opcode.OpConstant, []int{1}},
				{opcode.OpPop, []int{}},
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code{
				{opcode.OpConstant, []int{0}},
				{opcode.OpConstant, []int{1}},
				{opcode.OpSub, []int{}},
				{opcode.OpPop, []int{}},
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code{
				{opcode.OpConstant, []int{0}},
				{opcode.OpConstant, []int{1}},
				{opcode.OpMul, []int{}},
				{opcode.OpPop, []int{}},
			},
		},
		{
			input:             "2 / 1",
			expectedConstants: []any{2, 1},
			expectedInstructions: []code{
				{opcode.OpConstant, []int{0}},
				{opcode.OpConstant, []int{1}},
				{opcode.OpDiv, []int{}},
				{opcode.OpPop, []int{}},
			},
		},
		{
			input:             "-1",
			expectedConstants: []any{1},
			expectedInstructions: []code{
				{opcode.OpConstant, []int{0}},
				{opcode.OpMinus, []int{}},
				{opcode.OpPop, []int{}},
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if true { 10 }; 3333;",
			expectedConstants: []any{10, 3333},
			expectedInstructions: []code{
				{opcode.OpTrue, []int{}},            // 0000
				{opcode.OpJumpNotTruthy, []int{10}}, // 0001 (jump to 10? 1byte op + 2 bytes operand -> 3 bytes. Pos 4)
				// 0004
				{opcode.OpConstant, []int{0}}, // 10
				{opcode.OpJump, []int{11}},    // jump over null to 3333
				// 0010
				{opcode.OpNull, []int{}},
				// 0011
				{opcode.OpPop, []int{}},       // pop result of if
				{opcode.OpConstant, []int{1}}, // 3333
				{opcode.OpPop, []int{}},
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `func() { return 5 + 10 }`,
			expectedConstants: []any{
				5,
				10,
				[]code{
					{opcode.OpConstant, []int{0}},
					{opcode.OpConstant, []int{1}},
					{opcode.OpAdd, []int{}},
					{opcode.OpReturnValue, []int{}},
				},
			},
			expectedInstructions: []code{
				{opcode.OpClosure, []int{2, 0}}, // const 2 is fn, 0 free
				{opcode.OpPop, []int{}},
			},
		},
	}
	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testInstructions(
	expected []code,
	actual []byte,
) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q",
				i, concatted, actual)
		}
	}

	return nil
}

func concatInstructions(s []code) []byte {
	out := []byte{}

	for _, i := range s {
		out = append(out, opcode.Make(i.op, i.operands...)...)
	}

	return out
}

func testConstants(
	t *testing.T,
	expected []any,
	actual []object.Object,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. want=%d, got=%d",
			len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s",
					i, err)
			}
		case []code:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T",
					i, actual[i])
			}
			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - failed to match instructions: %s",
					i, err)
			}
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%d, got=%d",
			expected, result.Value)
	}

	return nil
}
