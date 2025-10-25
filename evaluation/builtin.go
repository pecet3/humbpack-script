package evaluation

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pecet3/hmbk-script/object"
)

var builtinFunctions map[string]*object.Builtin

func initBuiltInFunctions() {
	builtinFunctions = map[string]*object.Builtin{
		"loop": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				conditionFunc, ok1 := args[0].(*object.Function)
				bodyFunc, ok2 := args[1].(*object.Function)
				if !ok1 || !ok2 {
					return newError("arguments to `loop` must be functions")
				}

				env := object.NewClosedEnvironment(conditionFunc.Env)

				for {
					cond := Eval(conditionFunc.Body, env)
					if !isTruthy(cond) {
						break
					}
					if Eval(bodyFunc.Body, env).Type() != object.NULL {
						break
					}
				}
				return NULL
			},
		},
		"typeof": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				return &object.String{
					Value: strings.ToLower(fmt.Sprintf("%s", args[0].Type())),
				}
			},
		},
		"append": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return newError("wrong number of arguments. got=%d, min=2",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.Array:
					arg.Elements = append(arg.Elements, args[1:]...)
					return NULL
				case *object.Hash:
					if len(args) < 2 {
						return newError("wrong number of arguments. got=%d, min=3",
							len(args))
					}
					if key, ok := args[1].(object.Hashable); ok {
						pair := object.HashPair{
							Key:   args[1],
							Value: args[2],
						}
						arg.Pairs[key.HashKey()] = pair
						return NULL
					}
				default:
					return newError("first argument must be an array in push method")
				}
				return NULL
			},
		},
		"delete": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.Array:
					seeking := args[1]
					newElements := []object.Object{}
					for _, element := range arg.Elements {
						if element.Inspect() == seeking.Inspect() {
							continue
						}
						newElements = append(newElements, element)
					}
					arg.Elements = newElements
					return NULL
				case *object.Hash:
					if key, ok := args[1].(object.Hashable); ok {
						delete(arg.Pairs, key.HashKey())
					}
				default:
					return newError("first argument must be an array in push method")
				}
				return NULL
			},
		},
		"delete_index": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.Array:
					seeking, ok := args[1].(*object.Number)
					if !ok {
						return newError("second argument must be an integer")
					}
					if len(arg.Elements)-1 < int(seeking.Value) {
						return newError("provided index is too high. array has more than %d elements", seeking.Int()+1)
					}
					left := arg.Elements[:seeking.Int()]
					right := arg.Elements[seeking.Int()+1:]
					newArr := []object.Object{}
					newArr = append(newArr, left...)
					newArr = append(newArr, right...)
					arg.Elements = newArr
					return NULL
				default:
					return newError("first argument must be an array in push method")
				}
			},
		},
		"len": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.String:
					return &object.Number{Value: float64(len(arg.Value))}
				case *object.Array:
					return &object.Number{Value: float64(len(arg.Elements))}
				default:
					return newError("argument to `len` not supported, got %s",
						args[0].Type())
				}
			},
		},
		"print": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				for i, arg := range args {
					fmt.Print(arg.Inspect())
					if i == len(args)-1 {
						fmt.Println()
					}
				}
				return NULL
			},
		},
		"input": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.String:
					reader := bufio.NewReader(os.Stdin)
					line, err := reader.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							return newError("error reading input: %s", err.Error())
						}
					}
					line = strings.TrimRight(line, "\r\n")
					arg.Value = line
					return &object.String{Value: line}
				default:
					return newError("argument to `input` not supported, got %s",
						args[0].Type())
				}
			},
		},
		"bash": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.String:
					cmd := exec.Command("bash", "-c", arg.Inspect())
					if cmd.Err != nil {
						return newError("%s", cmd.Err.Error())
					}
					output, err := cmd.Output()
					if err != nil {
						return newError("%s", err.Error())
					}
					fmt.Println(string(output))
				default:
					return newError("argument to `bash` not supported, got %s",
						args[0].Type())
				}
				return NULL
			},
		},
		"to_string": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				return &object.String{
					Value: args[0].Inspect(),
				}
			},
		},
		"to_number": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				obj := args[0]
				if obj.Type() == object.STRING {
					val, err := strconv.ParseFloat(obj.Inspect(), 64)
					if err != nil {
						return newError("this string cannot be parset into float: %s",
							args[0].Inspect())
					}
					return &object.Number{
						Value: val,
					}
				}
				if obj.Type() == object.BOOL {
					b := obj.(*object.Bool)
					if b.Value {
						return &object.Number{
							Value: 1.0,
						}
					}
					return &object.Number{
						Value: 0.0,
					}
				}
				return newError("cannot convert value: %s to a number",
					args[0].Inspect())
			},
		},
	}

}
