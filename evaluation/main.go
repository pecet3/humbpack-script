package evaluation

import (
	"github.com/pecet3/hmbk-script/ast"
	"github.com/pecet3/hmbk-script/object"
)

var (
	TRUE  = &object.Bool{Value: true}
	FALSE = &object.Bool{Value: false}
	NULL  = &object.Null{}
)

func Eval(n ast.Node, env *object.Environment) object.Object {
	if len(builtinFunctions) == 0 {
		initBuiltInFunctions()
		initModules()
	}
	switch node := n.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.Module:
		modEnv := object.NewClosedEnvironment(env)
		evalProgram(node.Statements, modEnv)
		mod, ok := n.(*ast.Module)
		if !ok {
			return nil
		}
		env.SetModule(mod.Name, &object.Module{
			Name: mod.Name,
			Env:  modEnv,
		})
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Number{Value: float64(node.Value)}
	case *ast.FloatLiteral:
		return &object.Number{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isGlobalError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.Boolean:
		return boolToObject(node.Value)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isGlobalError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isGlobalError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isGlobalError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.MutStatement:
		val := Eval(node.Value, env)
		if isGlobalError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.ConstStatement:
		val := Eval(node.Value, env)
		if isGlobalError(val) {
			return val
		}
		if node.IsExport {
			env.SetPublicConst(node.Name.Value, val)
		}
		env.SetConst(node.Name.Value, val)

	case *ast.AssignmentStatement:
		val := Eval(node.Value, env)
		if isGlobalError(val) {
			return val
		}
		_, ok := env.Get(node.Name.Value)
		if !ok {
			return newGlobalError("assignment to undefined variable: %s", node.Name.Value)
		}
		isConst := env.IsConst(node.Name.Value)
		if isConst {
			return newGlobalError("assignment to const variable: %s", node.Name.Value)
		}
		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isGlobalError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isGlobalError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)

		if len(elements) == 1 && isGlobalError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isGlobalError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isGlobalError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.ModuleExpression:
		me := n.(*ast.ModuleExpression)
		bMod, ok := builtInModules[me.Left.String()]
		if ok {
			l := me.Index.Value
			val, ok := bMod.Env.Get(l)
			if !ok {
				return newGlobalError("buildin module %s has no symbol %s", me.Left.String(), l)
			}
			return val
		}

		left := Eval(node.Left, env)
		if isGlobalError(left) {
			return left
		}

		l := me.Index.Value
		mod := left.(*object.Module)

		val, ok := mod.Env.GetPublic(l)
		if !ok {
			if _, ok := mod.Env.GetMutNoOuter(l); ok {
				return newGlobalError("cannot access non-const symbol %s from module %s", l, mod.Name)
			}
			return newGlobalError("module %s has no public symbol %s", mod.Name, l)
		}
		return val
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return NULL
}

func boolToObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR
	}
	return false
}

func isGlobalError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.GLOBAL_ERROR
	}
	return false
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// expression

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newGlobalError("not a function: %s", fn.Type())
	}
}
func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewClosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// evals

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isGlobalError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)

		if !ok {
			return newGlobalError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isGlobalError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}

	}
	return &object.Hash{Pairs: pairs}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.NUMBER:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH:
		return evalHashIndexExpression(left, index)
	case left.Type() == object.MODULE:
		strIdx, ok := index.(*object.String)
		if !ok {
			return newGlobalError("module index must be string, got %s", index.Type())
		}
		mod := left.(*object.Module)
		val, ok := mod.Env.Get(strIdx.Value)
		if !ok {
			return newGlobalError("module %s has no symbol %s", mod.Name, strIdx.Value)
		}
		return val
	default:
		return newGlobalError("index operator must be an Number, not: %s", left.Type())
	}
}

func evalHashIndexExpression(left, index object.Object) object.Object {
	hashObject := left.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newGlobalError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalArrayIndexExpression(left, index object.Object) object.Object {
	array := left.(*object.Array)
	idx := index.(*object.Number).Int()
	max := int64(len(array.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return array.Elements[idx]
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isGlobalError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	mod, ok := env.GetModule(node.Value)
	if ok {
		return mod
	}
	val, ok := env.Get(node.Value)
	if ok {
		return val
	}
	if builtin, ok := builtinFunctions[node.Value]; ok {
		return builtin
	}

	return newGlobalError("identifier not found: " + node.Value)
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE || rt == object.ERROR {
				return result
			}
		}
	}
	return result
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isGlobalError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	}
	if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}
	return NULL
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusExpression(right)
	default:
		return newGlobalError("unknown operator: %s%s", operator, right.Type())
	}

}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusExpression(right object.Object) object.Object {
	if right.Type() != object.NUMBER {
		return newGlobalError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Number).Value
	return &object.Number{Value: -value}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.NUMBER && right.Type() == object.NUMBER:
		return evalNumberInfixExpression(operator, left, right)
	case operator == "==":
		if left.Type() == object.STRING && right.Type() == object.STRING {
			return boolToObject(left.Inspect() == right.Inspect())
		}
		return boolToObject(left == right)
	case operator == "!=":
		return boolToObject(left != right)
	case left.Type() == object.STRING && right.Type() == object.STRING:
		return evalStringsInfixExpression(operator, left, right)
	case left.Type() == object.NUMBER && right.Type() == object.STRING:
		left := left.(*object.Number)
		right := right.(*object.String)

		return evalStringAndNumberInfixExpression(operator, right, left)
	case left.Type() == object.STRING && right.Type() == object.NUMBER:
		left := left.(*object.String)
		right := right.(*object.Number)

		return evalStringAndNumberInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newGlobalError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newGlobalError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalNumberInfixExpression(
	operator string, left, right object.Object,
) object.Object {
	leftVal := left.(*object.Number).Value
	rightVal := right.(*object.Number).Value
	switch operator {
	case "+":
		return &object.Number{Value: leftVal + rightVal}
	case "-":
		return &object.Number{Value: leftVal - rightVal}
	case "*":
		return &object.Number{Value: leftVal * rightVal}
	case "/":
		return &object.Number{Value: leftVal / rightVal}
	case "<":
		return boolToObject(leftVal < rightVal)
	case ">":
		return boolToObject(leftVal > rightVal)
	case "==":
		return boolToObject(leftVal == rightVal)
	case "!=":
		return boolToObject(leftVal != rightVal)
	default:
		return newGlobalError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringsInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newGlobalError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalStringAndNumberInfixExpression(operator string, left *object.String, right *object.Number) object.Object {
	if operator != "+" {
		return newGlobalError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	return &object.String{Value: left.Value + right.Inspect()}
}
