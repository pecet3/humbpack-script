package vm

import (
	"fmt"

	"github.com/pecet3/hmbk-script/code"
	"github.com/pecet3/hmbk-script/compiler"
	"github.com/pecet3/hmbk-script/object"
)

const stackSize = 2048

type VM struct {
	consts       []object.Object
	stack        []object.Object
	sp           int
	ip           int
	instructions []byte
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		consts:       bytecode.Constants,
		stack:        make([]object.Object, stackSize),
		sp:           0,
		ip:           0,
		instructions: bytecode.Instructions,
	}
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

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}
	}
	return nil
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
	if vm.sp == 0 {
		return nil
	}
	vm.sp--
	o := vm.stack[vm.sp]
	vm.stack[vm.sp] = nil
	return o
}
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	fmt.Println(leftType, rightType)

	if leftType == object.NUMBER && rightType == object.NUMBER {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}
func (vm *VM) executeBinaryIntegerOperation(
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
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return vm.push(&object.Number{Value: result})
}
