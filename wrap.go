package cwrapmsg

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

type WrapData struct {
	line       int
	errName    string
	message    string
	parentFunc string
}

func FindWrapCalls(node *ast.File, fset *token.FileSet, fileName string) ([]WrapData, map[string]struct{}) {
	var (
		wrapDataSl    = []WrapData{}
		errorNamesMap = make(map[string]struct{})
		currentFunc   string
	)

	ast.Inspect(node, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.FuncDecl:
			currentFunc = expr.Name.Name
		case *ast.CallExpr:
			if fun, ok := expr.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "errors" &&
					(fun.Sel.Name == "Wrap" || fun.Sel.Name == "Wrapf") && len(expr.Args) >= 2 {

					message, isSecondArgumentString := getSecondArgumentName(expr)
					if !isSecondArgumentString {
						return false
					}

					position := fset.Position(expr.Pos())

					funcName, ok := isFirstArgumentFuncCall(expr)
					if ok && !isWrapMsgSuitable(funcName, message) {
						printIncorrectWrap(fileName, position.Line)
					} else {
						errName, ok := getFirstArgumentName(expr)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:       position.Line,
								errName:    errName,
								message:    message,
								parentFunc: currentFunc,
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
	if strings.Contains(funcName, "*ast.CallExpr") {
		funcName = getLastWord(funcName)
		message = getLastWord(message)

		return strings.ToLower(funcName) == strings.ToLower(message)
	}

	if !checkStandard(funcName, message) && !checkWithUnderScore(funcName, message) {
		return false
	}

	return true
}

func checkStandard(funcName, message string) bool {
	funcNameSl := strings.Split(funcName, ".")
	messageSl := strings.Split(cutMessage(message), ".")

	if len(messageSl) > len(funcNameSl) {
		return false
	}

	for idx := 0; idx < len(messageSl); idx++ {
		if strings.ToLower(messageSl[len(messageSl)-1-idx]) != strings.ToLower(funcNameSl[len(funcNameSl)-1-idx]) {
			return false
		}
	}

	return true
}

func checkWithUnderScore(funcName, message string) bool {
	return checkStandard(funcName, strings.ReplaceAll(message, "_", ""))
}

func getLastWord(str string) string {
	sl := strings.Split(str, ".")

	return sl[len(sl)-1]
}

func cutMessage(message string) string {
	result := make([]rune, 0, utf8.RuneCountInString(message))

	for _, v := range message {
		if unicode.IsLetter(v) || unicode.IsDigit(v) || v == '.' {
			result = append(result, v)
		} else {
			break
		}
	}

	return string(result)
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
