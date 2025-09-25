package evaluation

import (
	"github.com/pecet3/hmbk-script/ast"
	"github.com/pecet3/hmbk-script/object"
)

func Eval(n ast.Node) object.Object {
	switch node := n.(type) {
	case *ast.Program:
		return evalStatement(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	}
	return nil
}

func evalStatement(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)
	}
	return result
}
