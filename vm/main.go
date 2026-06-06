package vm

import (
	"fmt"
	"log"

	"github.com/pecet3/hmbk-script/code"
	"github.com/pecet3/hmbk-script/compiler"
	"github.com/pecet3/hmbk-script/object"
)

var True = &object.Bool{Value: true}
var False = &object.Bool{Value: false}
var Null = &object.Null{}

const (
	stackSize  = 2048
	GlobalSize = 65536
)

type VM struct {
	consts       []object.Object
	stack        []object.Object
	sp           int
	globals      []object.Object
	instructions []byte
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		consts:       bytecode.Constants,
		stack:        make([]object.Object, stackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalSize),
		instructions: bytecode.Instructions,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) LastPoppedStackElem() object.Object {
	fmt.Println(vm.stack[vm.sp])
	return vm.stack[vm.sp]
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {

	for ip := 0; ip < len(vm.instructions); ip++ {

		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.OpConstant:
			constIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			if constIndex >= len(vm.consts) {
				return fmt.Errorf("constant index out of range: %d", constIndex)
			}
			if err := vm.push(vm.consts[constIndex]); err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			if globalIndex >= GlobalSize {
				return fmt.Errorf("global index out of range: %d", globalIndex)
			}
			vm.globals[globalIndex] = vm.StackTop()
		case code.OpGetGlobal:
			globalIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements
			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}
	}
	return nil
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
func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}
func (vm *VM) push(o object.Object) error {
	if vm.sp >= stackSize {
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
func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Bool:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
func nativeBoolToBooleanObject(input bool) *object.Bool {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	if left.Type() == object.NUMBER && right.Type() == object.NUMBER {
		return vm.executeNumberComparison(op, left, right)
	}
	log.Println(right.Type(), left.Type())
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}
func (vm *VM) executeNumberComparison(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(*object.Number).Value
	rightValue := right.(*object.Number).Value
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}
func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()
	switch {
	case leftType == object.NUMBER && rightType == object.NUMBER:
		return vm.executeBinaryNumberOperation(op, left, right)
	case leftType == object.STRING && rightType == object.STRING:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}
func (vm *VM) executeBinaryStringOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeBinaryNumberOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(*object.Number).Value
	rightValue := right.(*object.Number).Value

	var result float64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		if rightValue == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown number operator: %d", op)
	}

	return vm.push(&object.Number{Value: result})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.BOOL:
		if operand == True {
			return vm.push(False)
		}
		return vm.push(True)
	case object.NULL:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.NUMBER:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}
func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Number).Int()
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
