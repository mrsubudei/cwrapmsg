package cwrapmsg

import (
	"go/ast"
	"go/token"
	"strings"
)

type WrapData struct {
	line    int
	errName string
	message string
}

func FindWrapCalls(node *ast.File, fset *token.FileSet, fileName string) ([]WrapData, map[string]struct{}) {
	wrapDataSl := []WrapData{}
	errorNamesMap := make(map[string]struct{})

	ast.Inspect(node, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "errors" &&
					(fun.Sel.Name == "Wrap" || fun.Sel.Name == "Wrapf") && len(callExpr.Args) >= 2 {

					message, isSecondArgumentString := getSecondArgumentName(callExpr)
					if !isSecondArgumentString {
						return false
					}

					position := fset.Position(callExpr.Pos())

					funcName, ok := isFirstArgumentFuncCall(callExpr)
					if ok && !isWrapMsgSuitable(funcName, message) {
						printIncorrectWrap(fileName, position.Line)
					} else {
						errName, ok := getFirstArgumentName(callExpr)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:    position.Line,
								errName: errName,
								message: message,
							})
						}
					}
				}
			}
		}

		return true
	})

	return wrapDataSl, errorNamesMap
}

func isWrapMsgSuitable(funcName, message string) bool {
	return strings.Contains(funcName, message)
}

func getFirstArgumentName(callExpr *ast.CallExpr) (string, bool) {
	if firstArg, ok := callExpr.Args[0].(*ast.Ident); ok {
		return firstArg.Name, true
	}

	return "", false
}

func getSecondArgumentName(callExpr *ast.CallExpr) (string, bool) {
	if strLit, ok := callExpr.Args[1].(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		return strings.Trim(strLit.Value, "\\\""), true
	}

	return "", false
}

func isFirstArgumentFuncCall(callExpr *ast.CallExpr) (string, bool) {
	if callExp, isFuncCall := callExpr.Args[0].(*ast.CallExpr); isFuncCall {
		return getFuncName(callExp.Fun), true
	}

	return "", false
}
