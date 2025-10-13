package evaluation

import (
	"github.com/pecet3/hmbk-script/object"
)

var builtinMathMap map[string]*object.Builtin

func moduleMathInit() map[string]*object.Builtin {
	builtinMathMap = map[string]*object.Builtin{
		"add": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2", len(args))
				}

				left, ok1 := args[0].(*object.Number)
				right, ok2 := args[1].(*object.Number)
				if !ok1 || !ok2 {
					return newError("arguments to `add` must be Numbers")
				}
				return &object.Number{Value: left.Value + right.Value}
			},
		},
	}
	return builtinMathMap
}
