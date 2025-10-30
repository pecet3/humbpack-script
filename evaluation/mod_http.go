package evaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pecet3/hmbk-script/object"
)

func ModHttp() *object.Environment {
	env := object.NewEnvironment()
	srv := http.NewServeMux()

	// -------------------------------
	// get_json(req)
	// -------------------------------
	env.SetConst("get_json", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newGlobalError("wrong number of arguments. got=%d, want=1", len(args))
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

				var parsed interface{}
				if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
					return newError("invalid JSON: %s", err)
				}
				return goValueToObject(parsed)
			default:
				return newGlobalError("argument to `get_json` not supported, got %s", args[0].Type())
			}
		},
	})

	// -------------------------------
	// get_param(req, key)
	// -------------------------------
	env.SetConst("get_param", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=2", len(args))
			}
			reqObj, ok := args[0].(*object.BuiltinObject)
			if !ok {
				return newError("first argument must be a request")
			}
			keyObj, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument must be a string key")
			}
			req, ok := reqObj.Value.(*http.Request)
			if !ok {
				return newError("wrong request type")
			}
			val := req.URL.Query().Get(keyObj.Value)
			return &object.String{Value: val}
		},
	})

	// -------------------------------
	// get_header(req, key)
	// -------------------------------
	env.SetConst("get_header", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=2", len(args))
			}
			reqObj, ok := args[0].(*object.BuiltinObject)
			if !ok {
				return newError("first argument must be a request")
			}
			keyObj, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument must be a string")
			}
			req, ok := reqObj.Value.(*http.Request)
			if !ok {
				return newError("wrong request type")
			}
			return &object.String{Value: req.Header.Get(keyObj.Value)}
		},
	})

	// -------------------------------
	// handle(path, fn)
	// -------------------------------
	env.SetConst("handle", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument must be string path")
			}
			fn, ok := args[1].(*object.Function)
			if !ok {
				return newError("second argument for handler should be a function")
			}

			srv.HandleFunc(path.Value, func(w http.ResponseWriter, r *http.Request) {
				fnEnv := fn.Env
				fnEnv.SetConst("req", &object.BuiltinObject{Value: r})
				fnEnv.SetConst("res", &object.BuiltinObject{Value: w})
				result := Eval(fn.Body, fnEnv)

				if isGlobalError(result) || result.Type() == object.NULL {
					return
				}
				w.Write([]byte(result.Inspect()))
				fn.Env = nil
			})
			return NULL
		},
	})

	// -------------------------------
	// listen(addr)
	// -------------------------------
	env.SetConst("listen", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newGlobalError("wrong number of arguments. got=%d, want=1", len(args))
			}
			addrObj, ok := args[0].(*object.String)
			if !ok {
				return newError("argument must be string")
			}
			if err := http.ListenAndServe(addrObj.Value, srv); err != nil {
				fmt.Printf("Error starting server: %s\n", err)
			}
			return NULL
		},
	})

	// -------------------------------
	// write_json(res, obj)
	// -------------------------------
	env.SetConst("write_json", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=2", len(args))
			}
			resObj, ok := args[0].(*object.BuiltinObject)
			if !ok {
				return newError("first argument must be response writer")
			}
			res, ok := resObj.Value.(http.ResponseWriter)
			if !ok {
				return newError("wrong response type")
			}
			data := objectToGoValue(args[1])
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return newError("json marshal error: %s", err)
			}
			res.Header().Set("Content-Type", "application/json")
			res.Write(jsonBytes)
			return NULL
		},
	})

	// -------------------------------
	// get(url)
	// -------------------------------
	env.SetConst("get", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newGlobalError("wrong number of arguments. got=%d, want=1", len(args))
			}
			urlObj, ok := args[0].(*object.String)
			if !ok {
				return newError("argument must be string")
			}
			resp, err := http.Get(urlObj.Value)
			if err != nil {
				return newError("GET error: %s", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("Read body error:%s", err)
			}
			return &object.String{Value: string(body)}
		},
	})

	// -------------------------------
	// post_json(url, obj)
	// -------------------------------
	env.SetConst("post_json", &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newGlobalError("wrong number of arguments. got=%d, want=2", len(args))
			}
			urlObj, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument must be url string")
			}
			data := objectToGoValue(args[1])
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return newError("json marshal error: %s", err)
			}
			resp, err := http.Post(urlObj.Value, "application/json", bytes.NewBuffer(jsonBytes))
			if err != nil {
				return newError("POST error: %s", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return &object.String{Value: string(body)}
		},
	})

	return env
}
