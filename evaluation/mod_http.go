package evaluation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pecet3/hmbk-script/object"
)

const (
	MUX = "MUX"
)

func ModHttp() *object.Environment {
	env := object.NewEnvironment()

	srv := http.NewServeMux()

	env.SetConst("get_json", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newGlobalError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.BuiltinObject:
				req, ok := arg.Value.(*http.Request)
				if !ok {
					return newError("wrong request type")
				}

				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					return newError("error reading body: %s", err)
				}
				defer req.Body.Close()

				var parsed map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
					return newError("invalid JSON: %s", err)
				}
				hash := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}

				for k, v := range parsed {
					keyObj := &object.String{Value: k}
					valueObj := goValueToObject(v)
					hash.Pairs[keyObj.HashKey()] = object.HashPair{
						Key:   keyObj,
						Value: valueObj,
					}
				}

				return hash
			default:
				return newGlobalError("argument to `len` not supported, got %s",
					args[0].Type())
			}

		},
	})

	env.SetConst("handle", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:

				fn, ok := args[1].(*object.Function)
				if !ok {
					return newGlobalError("second argument for handler should be a function")
				}

				srv.HandleFunc(arg.Value, func(w http.ResponseWriter, r *http.Request) {

					fnEnv := fn.Env

					fnEnv.SetConst("req", &object.BuiltinObject{
						Value: r,
					})

					result := Eval(fn.Body, fnEnv)

					w.Write([]byte(result.Inspect()))
				})
				return NULL
			default:
				return newGlobalError("argument to `len` not supported, got %s",
					args[0].Type())
			}

		},
	})

	env.SetConst("listen", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newGlobalError("wrong number of arguments. got=%d, want=1",
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
				return newGlobalError("argument to `len` not supported, got %s",
					args[0].Type())
			}

		},
	})

	env.SetConst("set", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=2",
					len(args))
			}

			key, ok := args[0].(*object.String)
			if !ok {
				return newGlobalError("first argument to `set` must be STRING, got=%s",
					args[0].Type())
			}

			env.Set(key.Value, args[1])
			return &object.String{Value: fmt.Sprintf("set %s = %s", key.Value, args[1].Inspect())}
		},
	})

	env.SetConst("get", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newGlobalError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			key, ok := args[0].(*object.String)
			if !ok {
				return newGlobalError("argument to `get` must be STRING, got=%s",
					args[0].Type())
			}

			if val, ok := env.Get(key.Value); ok {
				return val
			}

			return &object.Null{} // lub newGlobalError("key not found: %s", key.Value)
		},
	})
	return env
}
