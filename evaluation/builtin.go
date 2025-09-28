package evaluation

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pecet3/hmbk-script/object"
)

var builtinFunctions map[string]*object.Builtin

func initBulitInFunctions() {
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

				env := object.NewEnclosedEnvironment(conditionFunc.Env)

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
		"len": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1",
						len(args))
				}
				switch arg := args[0].(type) {
				case *object.String:
					return &object.Integer{Value: int64(len(arg.Value))}
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
	}

}
