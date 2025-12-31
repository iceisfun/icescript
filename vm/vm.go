package vm

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/opcode"
	"github.com/iceisfun/icescript/token"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

var (
	True  = object.True // Use shared instances from object package
	False = object.False
	Null  = object.NullObj
)

type VM struct {
	constants []object.Object
	stack     []object.Object
	sp        int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object

	frames      []*Frame
	framesIndex int

	symbolTable *compiler.SymbolTable
	lastPopped  object.Object

	rng    *rand.Rand
	output io.Writer

	ctxStore map[string]any
	ctxMu    sync.RWMutex
	mu       sync.Mutex

	printPrefix string
}

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int
}

func (vm *VM) Rand() *rand.Rand {
	return vm.rng
}

func (vm *VM) Now() time.Time {
	return time.Now()
}

func (vm *VM) Writer() io.Writer {
	return vm.output
}

func (vm *VM) SetOutput(w io.Writer) {
	vm.output = w
}

func (vm *VM) Get(k string) (any, bool) {
	vm.ctxMu.RLock()
	defer vm.ctxMu.RUnlock()
	v, ok := vm.ctxStore[k]
	return v, ok
}

func (vm *VM) Set(k string, v any) {
	vm.ctxMu.Lock()
	defer vm.ctxMu.Unlock()
	vm.ctxStore[k] = v
}

func (vm *VM) SetPrintPrefix(p string) {
	vm.printPrefix = p
}

func (vm *VM) PrintPrefix() string {
	return vm.printPrefix
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions, SourceMap: bytecode.SourceMap, Name: "main"}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		globals:     make([]object.Object, GlobalSize),
		stack:       make([]object.Object, StackSize),
		sp:          0,
		frames:      frames,
		framesIndex: 1,
		symbolTable: bytecode.SymbolTable,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		output:      os.Stdout,
		ctxStore:    make(map[string]any),
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{cl: cl, ip: -1, basePointer: basePointer}
}

func (vm *VM) Instructions() []byte {
	return vm.currentFrame().cl.Fn.Instructions
}

// Invoke calls the given function/closure with arguments.
// It is cancellable via ctx.
func (vm *VM) Invoke(ctx context.Context, fn object.Object, args ...object.Object) (object.Object, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// 1. Validate function type
	closure, ok := fn.(*object.Closure)
	if !ok {
		return nil, fmt.Errorf("Invoke expected a function/closure, got %s", fn.Type())
	}

	// 2. Validate Arity
	if len(args) != closure.Fn.NumParameters {
		return nil, fmt.Errorf("wrong number of arguments: want=%d, got=%d", closure.Fn.NumParameters, len(args))
	}

	// 3. Prepare Stack
	vm.sp = 0

	// 4. Push Closure & Args
	err := vm.push(closure)
	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		err := vm.push(arg)
		if err != nil {
			return nil, err
		}
	}

	// 5. Setup Frame
	// In OpCall, basePointer points to the first argument.
	// vm.sp currently points after the last argument.
	// So basePointer should be vm.sp - len(args).
	basePointer := vm.sp - len(args)
	frame := NewFrame(closure, basePointer)
	vm.frames[0] = frame
	vm.framesIndex = 1

	// Set SP to reserve space for locals
	vm.sp = frame.basePointer + closure.Fn.NumLocals

	// 6. Run
	err = vm.run(ctx)
	if err != nil {
		return nil, err
	}

	// 7. Get Return Value
	return vm.lastPopped, nil
}

func (vm *VM) GetGlobal(name string) (object.Object, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.symbolTable == nil {
		return nil, fmt.Errorf("no symbol table available")
	}

	symbol, ok := vm.symbolTable.Resolve(name)
	if !ok {
		return nil, fmt.Errorf("undefined global: %s", name)
	}

	if symbol.Scope == compiler.GlobalScope {
		return vm.globals[symbol.Index], nil
	}

	return nil, fmt.Errorf("%s is not a global (scope: %s)", name, symbol.Scope)
}

func (vm *VM) Run(ctx context.Context) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return vm.run(ctx)
}

func (vm *VM) run(ctx context.Context) error {
	var (
		ip  int
		ins []byte
		op  opcode.Opcode
	)

	// Instruction counter for preemptive context check
	opsCount := 0

	for vm.currentFrame().ip < len(vm.Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.Instructions()
		op = opcode.Opcode(ins[ip])

		opsCount++
		if opsCount >= 1024 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				opsCount = 0
			}
		}

		switch op {
		case opcode.OpConstant:
			constIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpPop:
			vm.pop()

		case opcode.OpDup:
			err := vm.push(vm.StackTop())
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpAdd, opcode.OpSub, opcode.OpMul, opcode.OpDiv, opcode.OpMod:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
		case opcode.OpEqual, opcode.OpNotEqual, opcode.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
		case opcode.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpTrue:
			err := vm.push(True)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
		case opcode.OpFalse:
			err := vm.push(False)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
		case opcode.OpNull:
			err := vm.push(Null)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpJump:
			pos := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case opcode.OpJumpNotTruthy:
			pos := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			result, err := isTruthy(condition)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
			if !result {
				vm.currentFrame().ip = pos - 1
			}

		case opcode.OpSetGlobal:
			globalIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = unwrapTuple(vm.pop())

		case opcode.OpGetGlobal:
			globalIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpSetLocal:
			localIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = unwrapTuple(vm.pop())

		case opcode.OpGetLocal:
			localIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpArray:
			numElements := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if vm.sp-numElements < 0 {
				return vm.newRuntimeError("stack underflow in OpArray")
			}

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpHash:
			numElements := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if vm.sp-numElements < 0 {
				return vm.newRuntimeError("stack underflow in OpHash")
			}

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpSlice:
			end := vm.pop()
			start := vm.pop()
			left := vm.pop()

			err := vm.executeSliceExpression(left, start, end)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpCall:
			numArgs := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			// Callee is on stack before args
			callee := vm.stack[vm.sp-1-int(numArgs)]

			switch callee := callee.(type) {
			case *object.Closure:
				if int(numArgs) != callee.Fn.NumParameters {
					return vm.newRuntimeError("wrong number of arguments: want=%d, got=%d", callee.Fn.NumParameters, numArgs)
				}
				frame := NewFrame(callee, vm.sp-int(numArgs))
				err := vm.pushFrame(frame)
				if err != nil {
					return vm.newRuntimeError("%s", err.Error())
				}
				vm.sp = frame.basePointer + callee.Fn.NumLocals // Reserve space? Or stack grows dynamically?

			case *object.Builtin:
				args := vm.stack[vm.sp-int(numArgs) : vm.sp] // Get args slice
				result := callee.Fn(vm, args...)
				vm.sp = vm.sp - int(numArgs) - 1 // Pop args and function
				if result != nil {
					if rtErr, ok := result.(*object.Panic); ok {
						return vm.newRuntimeError("%s", rtErr.Message)
					}
					if crit, ok := result.(*object.Critical); ok {
						return vm.newRuntimeError("%s", crit.Message)
					}
					vm.push(result)
				} else {
					vm.push(Null)
				}

			default:
				return vm.newRuntimeError("calling non-function")
			}

		case opcode.OpReturnValue:
			returnValue := vm.pop()
			vm.lastPopped = returnValue

			if vm.framesIndex == 1 {
				// Returning from main frame
				vm.popFrame()
				return nil
			}

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1 // -1 to pop the function/closure itself

			err := vm.push(returnValue)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpReturn:
			vm.lastPopped = Null
			if vm.framesIndex == 1 {
				vm.popFrame()
				return nil
			}

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpClosure:
			constIndex := opcode.ReadUint16(ins[ip+1:])
			numFree := opcode.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			if vm.sp-int(numFree) < 0 {
				return vm.newRuntimeError("stack underflow in OpClosure")
			}

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpGetFree:
			freeIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpGetBuiltin:
			builtinIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return vm.newRuntimeError("%s", err.Error())
			}

		case opcode.OpDestructure:
			numElements := int(opcode.ReadUint8(ins[ip+1:]))
			vm.currentFrame().ip += 1

			obj := vm.pop()

			if tuple, ok := obj.(*object.Tuple); ok {
				// Check for checks (Case D)
				if len(tuple.Elements) < numElements {
					return vm.newRuntimeError("not enough values to unpack: have %d, want %d", len(tuple.Elements), numElements)
				}
				// Push elements in order so they can be popped in reverse order by subsequent Sets
				// Example: var a, b = fn() -> fn returns (1, 2)
				// Stack: [..., 1, 2]
				// OpSetLocal b -> pops 2
				// OpSetLocal a -> pops 1
				for i := 0; i < numElements; i++ {
					err := vm.push(tuple.Elements[i])
					if err != nil {
						return vm.newRuntimeError("%s", err.Error())
					}
				}
			} else {
				// Scalar value, treat as single element
				if numElements != 1 {
					return vm.newRuntimeError("cannot destructure non-tuple into %d values", numElements)
				}
				err := vm.push(obj)
				if err != nil {
					return vm.newRuntimeError("%s", err.Error())
				}
			}
		}
	}
	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

type RuntimeError struct {
	Message string
	Stack   []StackFrameInfo
}

func (e *RuntimeError) Error() string {
	var msg string
	msg = fmt.Sprintf("Runtime error: %s\n", e.Message)
	msg += "Stack trace:\n"
	for _, f := range e.Stack {
		msg += fmt.Sprintf("  at %s (%s:%d)\n", f.FunctionName, "script.ice", f.Line)
	}
	return msg
}

type StackFrameInfo struct {
	FunctionName string
	FileName     string
	Line         int
}

func (vm *VM) newRuntimeError(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)

	currentFrame := vm.currentFrame()

	// Default values
	line := 0
	fileName := "script.ice" // Default filename
	functionName := ""

	if currentFrame != nil {
		if currentFrame.cl != nil && currentFrame.cl.Fn != nil {
			// Resolve line number from SourceMap
			// ip is usually pointing to next instruction, so we look back
			rawIP := currentFrame.ip // Points to OpCode

			line = translateIPToLine(currentFrame.cl.Fn.SourceMap, rawIP)

			functionName = currentFrame.cl.Fn.Name
		}
	}

	// Capture standardized stack trace
	stackTrace := []string{}
	// Walk frames?
	for i := vm.framesIndex - 1; i >= 0; i-- { // framesIndex points to next empty slot, so framesIndex-1 is current top frame
		f := vm.frames[i]
		if f.cl != nil && f.cl.Fn != nil {
			fname := f.cl.Fn.Name
			if fname == "" {
				fname = "anonymous"
			}
			stackTrace = append(stackTrace, fmt.Sprintf("%s (line %d)", fname, translateIPToLine(f.cl.Fn.SourceMap, f.ip)))
		}
	}

	return &token.ScriptError{
		Kind:       token.ErrorKindRuntime,
		Message:    msg,
		Line:       line,
		File:       fileName,
		Function:   functionName,
		StackTrace: stackTrace,
	}
}

func translateIPToLine(sourceMap map[int]int, ip int) int {
	// Search backwards from the current IP to find the instruction start
	// ip points to the *next* instruction or the middle of current one depending on error context.
	// We check a small window backwards.
	for i := 0; i < 10; i++ {
		if l, ok := sourceMap[ip-i]; ok {
			return l
		}
	}
	return 0
}

func (vm *VM) stackTrace() []StackFrameInfo {
	var stack []StackFrameInfo

	// Iterate backwards from current frame
	for i := vm.framesIndex - 1; i >= 0; i-- {
		frame := vm.frames[i]
		if frame.cl == nil {
			continue
		}

		line := 0
		if frame.ip >= 0 {
			line = translateIPToLine(frame.cl.Fn.SourceMap, frame.ip)
		}

		name := frame.cl.Fn.Name
		if name == "" {
			name = "anonymous"
		}

		info := StackFrameInfo{
			FunctionName: name,
			FileName:     "script.ice",
			Line:         line,
		}
		stack = append(stack, info)
	}
	return stack
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) executeBinaryOperation(op opcode.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}
	// floats, strings...

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(
	op opcode.Opcode,
	left, right object.Object,
) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case opcode.OpAdd:
		return vm.push(&object.Integer{Value: leftVal + rightVal})
	case opcode.OpSub:
		return vm.push(&object.Integer{Value: leftVal - rightVal})
	case opcode.OpMul:
		return vm.push(&object.Integer{Value: leftVal * rightVal})
	case opcode.OpDiv:
		return vm.push(&object.Integer{Value: leftVal / rightVal})
	case opcode.OpMod:
		return vm.push(&object.Integer{Value: leftVal % rightVal})
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
}

func (vm *VM) executeComparison(op opcode.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ {
		return vm.executeFloatComparison(op, left, right)
	}

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeStringComparison(op, left, right)
	}

	// Handle primitive equality fallback
	if isPrimitive(left.Type()) && isPrimitive(right.Type()) {
		switch op {
		case opcode.OpEqual:
			return vm.push(nativeBoolToBooleanObject(right == left))
		case opcode.OpNotEqual:
			return vm.push(nativeBoolToBooleanObject(right != left))
		default:
			return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
		}
	}

	// For non-primitives, we require strict type matching and explicit support
	if left.Type() != right.Type() {
		return fmt.Errorf("type mismatch: %s %s %s", left.Type(), opString(op), right.Type())
	}

	// Check for explicit equality support
	if eqObj, ok := left.(object.ObjectEqual); ok {
		equal, err := eqObj.Equal(right)
		if err != nil {
			return err
		}
		if op == opcode.OpNotEqual {
			return vm.push(nativeBoolToBooleanObject(!equal))
		}
		return vm.push(nativeBoolToBooleanObject(equal))
	}

	return fmt.Errorf("equality not supported for type: %s", left.Type())
}

func isPrimitive(t object.ObjectType) bool {
	switch t {
	case object.INTEGER_OBJ, object.FLOAT_OBJ, object.BOOLEAN_OBJ, object.NULL_OBJ, object.STRING_OBJ:
		return true
	default:
		return false
	}
}

func opString(op opcode.Opcode) string {
	switch op {
	case opcode.OpEqual:
		return "=="
	case opcode.OpNotEqual:
		return "!="
	case opcode.OpGreaterThan:
		return ">"
	default:
		return fmt.Sprintf("OP(%d)", op)
	}
}

func (vm *VM) executeFloatComparison(
	op opcode.Opcode,
	left, right object.Object,
) error {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch op {
	case opcode.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case opcode.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	case opcode.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	default: // Support less than via reordering or future opcode
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeStringComparison(
	op opcode.Opcode,
	left, right object.Object,
) error {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case opcode.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case opcode.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeIntegerComparison(
	op opcode.Opcode,
	left, right object.Object,
) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case opcode.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case opcode.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	case opcode.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

// Helper to unwrap Tuple to first element (Case B)
func unwrapTuple(obj object.Object) object.Object {
	if tuple, ok := obj.(*object.Tuple); ok {
		return tuple.Elements[0]
	}
	return obj
}

func (vm *VM) executeBangOperator() error {
	operand := unwrapTuple(vm.pop())

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	switch operand.Type() {
	case object.INTEGER_OBJ:
		value := operand.(*object.Integer).Value
		return vm.push(&object.Integer{Value: -value})
	case object.FLOAT_OBJ:
		value := operand.(*object.Float).Value
		return vm.push(&object.Float{Value: -value})
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := 0; i < endIndex-startIndex; i++ {
		elements[i] = vm.stack[startIndex+i]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeSliceExpression(left, start, end object.Object) error {
	if left.Type() != object.ARRAY_OBJ {
		return fmt.Errorf("slice operator not supported: %s", left.Type())
	}

	arrayObject := left.(*object.Array)
	elements := arrayObject.Elements
	length := int64(len(elements))

	var startIndex int64 = 0
	var endIndex int64 = length

	if start != Null {
		if start.Type() != object.INTEGER_OBJ {
			return fmt.Errorf("slice start index must be INTEGER, got %s", start.Type())
		}
		startIndex = start.(*object.Integer).Value
	}

	if end != Null {
		if end.Type() != object.INTEGER_OBJ {
			return fmt.Errorf("slice end index must be INTEGER, got %s", end.Type())
		}
		endIndex = end.(*object.Integer).Value
	}

	// Adjust bounds
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > length {
		endIndex = length
	}
	if startIndex > endIndex {
		startIndex = endIndex
	}

	newElements := make([]object.Object, endIndex-startIndex)
	copy(newElements, elements[startIndex:endIndex])

	return vm.push(&object.Array{Elements: newElements})
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) error {
	if vm.framesIndex >= MaxFrames {
		return fmt.Errorf("stack overflow")
	}
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
	return nil
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree // pop free args

	closure := &object.Closure{Fn: function, Free: free}
	return vm.push(closure)
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj object.Object) (bool, error) {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value, nil
	case *object.Null:
		return false, nil
	case *object.Integer:
		return obj.Value != 0, nil
	case *object.Float:
		return obj.Value != 0.0, nil
	case *object.String:
		return obj.Value != "", nil
	default:
		return false, fmt.Errorf("condition must be boolean, got %s", obj.Type())
	}
}

func (vm *VM) SetGlobal(index int, val object.Object) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if index >= len(vm.globals) {
		return fmt.Errorf("global index %d out of bounds", index)
	}
	vm.globals[index] = val
	return nil
}
