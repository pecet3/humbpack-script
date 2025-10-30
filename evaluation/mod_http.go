package evaluation

import (
	"fmt"
	"net/http"

	"github.com/pecet3/hmbk-script/object"
)

const (
	MUX = "MUX"
)

func ModHttp() *object.Environment {
	env := object.NewEnvironment()

	srv := http.NewServeMux()

	env.SetConst("handle", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:

				fn, ok := args[1].(*object.Function)
				if !ok {
					return newError("second argument for handler should be a function")
				}

				srv.HandleFunc(arg.Value, func(w http.ResponseWriter, r *http.Request) {
					env.SetConst("req_host", &object.String{
						Value: r.Host,
					})

					fnEnv := object.NewClosedEnvironment(fn.Env)
					result := Eval(fn.Body, fnEnv)

					w.Write([]byte(result.Inspect()))
				})
				return NULL
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}

		},
	})

	env.SetConst("listen", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				addr := arg.Value

				if err := http.ListenAndServe(addr, srv); err != nil {
					fmt.Printf("Error starting server: %s\n", err)
				}

				return NULL
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}

		},
	})

	env.SetConst("set", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}

			key, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `set` must be STRING, got=%s",
					args[0].Type())
			}

			env.Set(key.Value, args[1])
			return &object.String{Value: fmt.Sprintf("set %s = %s", key.Value, args[1].Inspect())}
		},
	})

	env.SetConst("get", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			key, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `get` must be STRING, got=%s",
					args[0].Type())
			}

			if val, ok := env.Get(key.Value); ok {
				return val
			}

			return &object.Null{} // lub newError("key not found: %s", key.Value)
		},
	})
	return env
}
