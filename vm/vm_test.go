package vm

import (
	"context"
	"fmt"
	"testing"

	"github.com/iceisfun/icescript/ast"
	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
)

type vmTestCase struct {
	input    string
	expected any
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if true { 10 }", 10},
		{"if true { 10 } else { 20 }", 10},
		{"if false { 10 } else { 20 }", 20},
		{"if 1 { 10 }", 10},
		{"if 1 < 2 { 10 }", 10},
		{"if 1 > 2 { 10 } else { 20 }", 20},
		{"if 1 > 2 { 10 }", Null},
		{"if false { 10 }", Null},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"var one = 1; one", 1},
		{"var one = 1; var two = 2; one + two", 3},
		{"var one = 1; var two = one + one; one + two", 3},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	// Not implemented OpSub/OpAdd for strings yet in VM
	// But constant should work
	// tests := []vmTestCase{
	// 	{`"monkey"`, "monkey"},
	// }
	// runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}
	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"{}", map[object.HashKey]int64{}},
		{"{1: 2, 2: 3}", map[object.HashKey]int64{
			(&object.Integer{Value: 1}).HashKey(): 2,
			(&object.Integer{Value: 2}).HashKey(): 3,
		}},
		{"{1 + 1: 2 * 2, 3 + 3: 4 * 4}", map[object.HashKey]int64{
			(&object.Integer{Value: 2}).HashKey(): 4,
			(&object.Integer{Value: 6}).HashKey(): 16,
		}},
	}
	runVmTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var fivePlusTen = func() { return 5 + 10; };
			fivePlusTen();
			`,
			15,
		},
		{
			`
			var one = func() { 1; };
			var two = func() { 2; };
			one() + two()
			`,
			3,
		},
		{
			`
			var a = func() { 1 };
			var b = func() { a() + 1 };
			var c = func() { b() + 1 };
			c();
			`,
			3,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionsTerminating(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var earlyExit = func() { return 99; 100; };
			earlyExit();
			`,
			99,
		},
		{
			`
			var earlyExit = func() { return 99; return 100; };
			earlyExit();
			`,
			99,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var identity = func(a) { a; };
			identity(4);
			`,
			4,
		},
		{
			`
			var sum = func(a, b) { a + b; };
			sum(1, 2);
			`,
			3,
		},
	}
	runVmTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var newClosure = func(a) {
				return func() { a; };
			};
			var closure = newClosure(99);
			closure();
			`,
			99,
		},
		{
			`
			var makeAdder = func(base) {
				return func(a) { base + a; };
			};
			var addTwo = makeAdder(2);
			addTwo(2);
			`,
			4,
		},
	}
	runVmTests(t, tests)
}

func TestBuiltinRandomItem(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var a = [];
			randomItem(a);
			`,
			Null,
		},
		{
			`
			var a = [1];
			randomItem(a);
			`,
			1,
		},
		{
			`
			var a = [10, 20, 30];
			var x = randomItem(a);
			x == 10 || x == 20 || x == 30;
			`,
			true,
		},
	}
	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run(context.Background())
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testExpectedObject(t *testing.T, expected any, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case *object.Null:
		if actual != Null {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not Array: %T (%+v)", actual, actual)
			return
		}
		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d",
				len(expected), len(array.Elements))
			return
		}
		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("element %d wrong: %s", i, err)
			}
		}
	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object not Hash: %T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of pairs. want=%d, got=%d",
				len(expected), len(hash.Pairs))
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key")
			}

			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("pair value wrong: %s", err)
			}
		}
	}
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

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%t, got=%t",
			expected, result.Value)
	}

	return nil
}
