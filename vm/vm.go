package vm

import (
	"context"
	"fmt"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/opcode"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	// Call frames
	frames      []*Frame
	framesIndex int
}

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
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

func (vm *VM) Run(ctx context.Context) error {
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
				return err
			}

		case opcode.OpPop:
			vm.pop()

		case opcode.OpAdd, opcode.OpSub, opcode.OpMul, opcode.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case opcode.OpEqual, opcode.OpNotEqual, opcode.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case opcode.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case opcode.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case opcode.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case opcode.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case opcode.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case opcode.OpJump:
			pos := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case opcode.OpJumpNotTruthy:
			pos := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case opcode.OpSetGlobal:
			globalIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()

		case opcode.OpGetGlobal:
			globalIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case opcode.OpSetLocal:
			localIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()

		case opcode.OpGetLocal:
			localIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}

		case opcode.OpArray:
			numElements := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if vm.sp-numElements < 0 {
				return fmt.Errorf("stack underflow in OpArray")
			}

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}

		case opcode.OpHash:
			numElements := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if vm.sp-numElements < 0 {
				return fmt.Errorf("stack underflow in OpHash")
			}

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}

		case opcode.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case opcode.OpSlice:
			end := vm.pop()
			start := vm.pop()
			left := vm.pop()

			err := vm.executeSliceExpression(left, start, end)
			if err != nil {
				return err
			}

		case opcode.OpCall:
			numArgs := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			// Callee is on stack before args
			callee := vm.stack[vm.sp-1-int(numArgs)]

			switch callee := callee.(type) {
			case *object.Closure:
				if int(numArgs) != callee.Fn.NumParameters {
					return fmt.Errorf("wrong number of arguments: want=%d, got=%d", callee.Fn.NumParameters, numArgs)
				}
				frame := NewFrame(callee, vm.sp-int(numArgs))
				err := vm.pushFrame(frame)
				if err != nil {
					return err
				}
				vm.sp = frame.basePointer + callee.Fn.NumLocals // Reserve space? Or stack grows dynamically?
				// Stack grows on push. But local vars slots need to be logically mapped.
				// With OpSetLocal we assign to vm.stack[base + index].
				// If index > numArgs, we need to ensure stack is large enough?
				// The VM implementation usually assumes "locals" are reserved or we just index into stack which is pre-allocated.
				// But we need to make sure sp is moved past the locals.
				// In this book-based design, locals include params.
				// So if we have 2 params and 3 vars, sp should point to base + 5?
				// OpSetLocal writes to index. OpGetLocal reads.
				// If we declare a local, we usually emit code to put nil or the value on the stack via logic (e.g. VarStatement emits code).
				// So standard stack machine behavior is fine. The SP grows as we push locals.

			case *object.Builtin:
				args := vm.stack[vm.sp-int(numArgs) : vm.sp] // Get args slice
				result := callee.Fn(args...)
				vm.sp = vm.sp - int(numArgs) - 1 // Pop args and function
				if result != nil {
					vm.push(result)
				} else {
					vm.push(Null)
				}

			default:
				return fmt.Errorf("calling non-function")
			}

		case opcode.OpReturnValue:
			returnValue := vm.pop()

			if vm.framesIndex == 1 {
				// Returning from main frame
				vm.popFrame()
				return nil
			}

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1 // -1 to pop the function/closure itself

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case opcode.OpReturn:
			if vm.framesIndex == 1 {
				vm.popFrame()
				return nil
			}

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}

		case opcode.OpClosure:
			constIndex := opcode.ReadUint16(ins[ip+1:])
			numFree := opcode.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			if vm.sp-int(numFree) < 0 {
				return fmt.Errorf("stack underflow in OpClosure")
			}

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}

		case opcode.OpGetFree:
			freeIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}

		case opcode.OpGetBuiltin:
			builtinIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
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

	switch op {
	case opcode.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case opcode.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
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

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

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

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
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

func isTruthy(obj object.Object) bool {
	switch obj {
	case Null:
		return false
	case True:
		return true
	case False:
		return false
	default:
		return true
	}
}

func (vm *VM) SetGlobal(index int, val object.Object) error {
	if index >= len(vm.globals) {
		return fmt.Errorf("global index %d out of bounds", index)
	}
	vm.globals[index] = val
	return nil
}

func (vm *VM) GetGlobal(index int) (object.Object, error) {
	if index >= len(vm.globals) {
		return nil, fmt.Errorf("global index %d out of bounds", index)
	}
	return vm.globals[index], nil
}
