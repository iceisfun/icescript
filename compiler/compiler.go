package compiler

import (
	"fmt"

	"github.com/iceisfun/icescript/ast"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/opcode"
)

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
	lastLine   int

	symbolDefinitions map[ast.Node][]Symbol
}

type CompilationScope struct {
	instructions        []byte
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	sourceMap           map[int]int // instruction index -> line number
}

type EmittedInstruction struct {
	Opcode   opcode.Opcode
	Position int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        []byte{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		sourceMap:           make(map[int]int),
	}

	symbolTable := NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:         []object.Object{},
		symbolTable:       symbolTable,
		symbolDefinitions: make(map[ast.Node][]Symbol),
		scopes:            []CompilationScope{mainScope},
		scopeIndex:        0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		c.scanSymbols(node.Statements)
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		c.lastLine = node.Token.Line
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(opcode.OpPop)

	case *ast.InfixExpression:
		c.lastLine = node.Token.Line
		if node.Operator == "<" {
			err := c.Compile(node.Right) // Reorder for <
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(opcode.OpGreaterThan)
			return nil
		}

		if node.Operator == "<=" {
			// a <= b  <=>  not (a > b)
			// a > b
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}
			err = c.Compile(node.Right)
			if err != nil {
				return err
			}
			c.emit(opcode.OpGreaterThan)
			c.emit(opcode.OpBang) // Not
			return nil
		}

		if node.Operator == ">=" {
			// a >= b  ->  !(a < b) -> !(b > a)

			// b > a
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(opcode.OpGreaterThan)
			c.emit(opcode.OpBang)
			return nil
		}

		if node.Operator == "&&" {
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			// Stack: [left]
			c.emit(opcode.OpDup)
			// Stack: [left, left]
			jumpPos := c.emit(opcode.OpJumpNotTruthy, 9999)

			// Stack if true: [left] (we need to pop it before right)
			c.emit(opcode.OpPop)
			// Stack: []

			err = c.Compile(node.Right)
			if err != nil {
				return err
			}

			afterRightPos := len(c.currentInstructions())
			c.changeOperand(jumpPos, afterRightPos)

			return nil
		}

		if node.Operator == "||" {
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			// Stack: [left]
			c.emit(opcode.OpDup)
			// Stack: [left, left]
			jumpNotTruthyPos := c.emit(opcode.OpJumpNotTruthy, 9999)

			// Left was truthy, so Short-circuit!
			// Stack: [left]
			jumpToEndPos := c.emit(opcode.OpJump, 9999)

			// EvalRight: Left was falsy
			afterLeftPos := len(c.currentInstructions())
			c.changeOperand(jumpNotTruthyPos, afterLeftPos)

			// Stack: [left] -> Pop it
			c.emit(opcode.OpPop)

			err = c.Compile(node.Right)
			if err != nil {
				return err
			}

			afterRightPos := len(c.currentInstructions())
			c.changeOperand(jumpToEndPos, afterRightPos)

			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(opcode.OpAdd)
		case "-":
			c.emit(opcode.OpSub)
		case "*":
			c.emit(opcode.OpMul)
		case "/":
			c.emit(opcode.OpDiv)
		case "%":
			c.emit(opcode.OpMod)
		case ">":
			c.emit(opcode.OpGreaterThan)
		case "==":
			c.emit(opcode.OpEqual)
		case "!=":
			c.emit(opcode.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		c.lastLine = node.Token.Line
		integer := &object.Integer{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(integer))

	case *ast.FloatLiteral:
		c.lastLine = node.Token.Line
		f := &object.Float{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(f))

	case *ast.Boolean:
		c.lastLine = node.Token.Line
		if node.Value {
			c.emit(opcode.OpTrue)
		} else {
			c.emit(opcode.OpFalse)
		}

	case *ast.PrefixExpression:
		c.lastLine = node.Token.Line
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(opcode.OpBang)
		case "-":
			c.emit(opcode.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		c.lastLine = node.Token.Line
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit JumpNotTruthy with a placeholder offset
		jumpNotTruthyPos := c.emit(opcode.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(opcode.OpPop) {
			c.removeLastPop()
		}

		// Emit Jump with placeholder
		jumpPos := c.emit(opcode.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(opcode.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(opcode.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		c.lastLine = node.Token.Line
		symbols, ok := c.symbolDefinitions[node]
		if !ok {
			symbols = make([]Symbol, len(node.Names))
			for i, name := range node.Names {
				symbols[i] = c.symbolTable.Define(name.Value)
			}
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		// If multiple names, emit Destructure
		if len(node.Names) > 1 {
			c.emit(opcode.OpDestructure, len(node.Names))
		}

		// Assign in reverse order (stack is LIFO)
		for i := len(symbols) - 1; i >= 0; i-- {
			symbol := symbols[i]
			if symbol.Scope == GlobalScope {
				c.emit(opcode.OpSetGlobal, symbol.Index)
			} else {
				c.emit(opcode.OpSetLocal, symbol.Index)
			}
		}

	case *ast.ShortVarDeclaration:
		c.lastLine = node.Token.Line
		symbols, ok := c.symbolDefinitions[node]
		if !ok {
			symbols = make([]Symbol, len(node.Names))
			for i, name := range node.Names {
				symbols[i] = c.symbolTable.Define(name.Value)
			}
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if len(node.Names) > 1 {
			c.emit(opcode.OpDestructure, len(node.Names))
		}

		for i := len(symbols) - 1; i >= 0; i-- {
			symbol := symbols[i]
			if symbol.Scope == GlobalScope {
				c.emit(opcode.OpSetGlobal, symbol.Index)
			} else {
				c.emit(opcode.OpSetLocal, symbol.Index)
			}
		}

	case *ast.AssignExpression:
		c.lastLine = node.Token.Line
		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			return fmt.Errorf("variable %s not defined", node.Name.Value)
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpSetGlobal, symbol.Index)
			c.emit(opcode.OpGetGlobal, symbol.Index)
		} else if symbol.Scope == LocalScope {
			c.emit(opcode.OpSetLocal, symbol.Index)
			c.emit(opcode.OpGetLocal, symbol.Index)
		} else {
			return fmt.Errorf("assignment to %s not supported", symbol.Scope)
		}

	case *ast.Identifier:
		c.lastLine = node.Token.Line
		symbol, ok := c.symbolTable.Resolve(node.Value)

		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpGetGlobal, symbol.Index)
		} else if symbol.Scope == LocalScope {
			c.emit(opcode.OpGetLocal, symbol.Index)
		} else if symbol.Scope == BuiltinScope {
			c.emit(opcode.OpGetBuiltin, symbol.Index)
		} else if symbol.Scope == FreeScope {
			c.emit(opcode.OpGetFree, symbol.Index)
		}

	case *ast.StringLiteral:
		c.lastLine = node.Token.Line
		str := &object.String{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		c.lastLine = node.Token.Line
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(opcode.OpArray, len(node.Elements))

	case *ast.MapLiteral:
		c.lastLine = node.Token.Line
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		// Sort keys for deterministic output? AST expressions are not easily sortable.
		// We'll just iterate in random order, but MapLiteral in AST is map[Expression]Expression.
		// To be deterministic compile-time, we should sort if we can.
		// For now, let's just loop.

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		c.lastLine = node.Token.Line
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(opcode.OpIndex)

	case *ast.SliceExpression:
		c.lastLine = node.Token.Line
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		if node.Start != nil {
			err = c.Compile(node.Start)
			if err != nil {
				return err
			}
		} else {
			c.emit(opcode.OpNull)
		}

		if node.End != nil {
			err = c.Compile(node.End)
			if err != nil {
				return err
			}
		} else {
			c.emit(opcode.OpNull)
		}

		c.emit(opcode.OpSlice)

	case *ast.FunctionLiteral:
		c.lastLine = node.Token.Line
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(opcode.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(opcode.OpReturnValue) {
			c.emit(opcode.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions, sourceMap := c.leaveScope()

		for _, s := range freeSymbols {
			// Emit code to load the free variables onto stack before creating closure
			if s.Scope == LocalScope {
				c.emit(opcode.OpGetLocal, s.Index)
			} else if s.Scope == FreeScope {
				c.emit(opcode.OpGetFree, s.Index)
			}
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
			SourceMap:     sourceMap,
			Name:          node.Name,
		}

		c.emit(opcode.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	case *ast.NullLiteral:
		c.lastLine = node.Token.Line
		c.emit(opcode.OpNull)

	case *ast.ReturnStatement:
		c.lastLine = node.Token.Line
		if node.ReturnValue == nil {
			c.emit(opcode.OpNull)
		} else {
			err := c.Compile(node.ReturnValue)
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpReturnValue)

	case *ast.CallExpression:
		c.lastLine = node.Token.Line
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpCall, len(node.Arguments))

	case *ast.ForStatement:
		c.lastLine = node.Token.Line
		// Init
		if node.Init != nil {
			err := c.Compile(node.Init)
			if err != nil {
				return err
			}
		}

		startPos := len(c.currentInstructions())

		var jumpNotTruthyPos int

		if node.Condition != nil {
			err := c.Compile(node.Condition)
			if err != nil {
				return err
			}
			// jumpNotTruthy to end
			jumpNotTruthyPos = c.emit(opcode.OpJumpNotTruthy, 9999)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Post (run after body, before jumping back)
		if node.Post != nil {
			err := c.Compile(node.Post)
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpJump, startPos)

		if node.Condition != nil {
			afterBodyPos := len(c.currentInstructions())
			c.changeOperand(jumpNotTruthyPos, afterBodyPos)
		}
	}

	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op opcode.Opcode, operands ...int) int {
	ins := opcode.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	// Record line number
	if c.lastLine > 0 {
		c.scopes[c.scopeIndex].sourceMap[pos] = c.lastLine
	}

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op opcode.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op opcode.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	newI := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = newI
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, opcode.Make(opcode.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = opcode.OpReturnValue
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := opcode.Opcode(c.currentInstructions()[opPos])
	newInstruction := opcode.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) currentInstructions() []byte {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        []byte{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		sourceMap:           make(map[int]int),
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() ([]byte, map[int]int) {
	instructions := c.currentInstructions()
	sourceMap := c.scopes[c.scopeIndex].sourceMap

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	return instructions, sourceMap
}

type Bytecode struct {
	Instructions []byte
	Constants    []object.Object
	SymbolTable  *SymbolTable
	SourceMap    map[int]int
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
		SymbolTable:  c.symbolTable,
		SourceMap:    c.scopes[c.scopeIndex].sourceMap,
	}
}

func (c *Compiler) SymbolTable() *SymbolTable {
	return c.symbolTable
}

func (c *Compiler) scanSymbols(statements []ast.Statement) {
	for _, s := range statements {
		switch s := s.(type) {
		case *ast.LetStatement:
			symbols := make([]Symbol, len(s.Names))
			for i, name := range s.Names {
				symbols[i] = c.symbolTable.Define(name.Value)
			}
			c.symbolDefinitions[s] = symbols
		case *ast.ShortVarDeclaration:
			symbols := make([]Symbol, len(s.Names))
			for i, name := range s.Names {
				symbols[i] = c.symbolTable.Define(name.Value)
			}
			c.symbolDefinitions[s] = symbols
		}
	}
}
