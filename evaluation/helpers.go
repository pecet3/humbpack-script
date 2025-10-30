package evaluation

import (
	"fmt"

	"github.com/pecet3/hmbk-script/object"
)

func newGlobalError(format string, a ...interface{}) *object.GlobalError {
	return &object.GlobalError{Message: fmt.Sprintf(format, a...)}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
func goValueToObject(v interface{}) object.Object {
	switch val := v.(type) {
	case string:
		return &object.String{Value: val}
	case float64:
		return &object.Number{Value: val}
	case bool:
		return &object.Bool{Value: val}
	case nil:
		return &object.Null{}
	case map[string]interface{}:
		h := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}
		for k, vv := range val {
			keyObj := &object.String{Value: k}
			valueObj := goValueToObject(vv)
			h.Pairs[keyObj.HashKey()] = object.HashPair{Key: keyObj, Value: valueObj}
		}
		return h
	case []interface{}:
		elements := []object.Object{}
		for _, elem := range val {
			elements = append(elements, goValueToObject(elem))
		}
		return &object.Array{Elements: elements}
	default:
		return &object.String{Value: fmt.Sprintf("%v", val)}
	}
}
