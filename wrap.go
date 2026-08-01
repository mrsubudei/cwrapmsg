package cwrapmsg

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

type WrapData struct {
	line          int
	errName       string
	message       string
	parentFunc    string
	nestingsLevel int
}

func FindWrapCalls(
	node *ast.File,
	fset *token.FileSet,
	fileName string,
	wrapPackageName string,
	checkCyrillic bool,
) ([]WrapData, map[string]struct{}) {

	wrapDataSl := []WrapData{}
	errorNamesMap := make(map[string]struct{})
	var currentFunc string
	nestingLevel := 0
	funcEndLines := Stack{}

	ast.Inspect(node, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.FuncDecl:
			position := fset.Position(expr.End())
			nestingLevel = 1
			funcEndLines = Stack{}
			funcEndLines.Add(position.Line)

			if expr.Recv != nil && len(expr.Recv.List) > 0 {
				receiverName := ""
				switch t := expr.Recv.List[0].Type.(type) {
				case *ast.Ident:
					receiverName = t.Name
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						receiverName = ident.Name
					}
				}
				currentFunc = receiverName + "_" + expr.Name.Name
			} else {
				currentFunc = expr.Name.Name
			}
		case *ast.FuncLit:
			nestingLevel++
			position := fset.Position(expr.End())
			funcEndLines.Add(position.Line)
		case *ast.CallExpr:
			if fun, ok := expr.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == wrapPackageName &&
					(fun.Sel.Name == "Wrap" || fun.Sel.Name == "Wrapf") && len(expr.Args) >= 2 {

					position := fset.Position(expr.Pos())
					currFuncEndLine := funcEndLines.Get()
					for position.Line > currFuncEndLine {
						nestingLevel--
						funcEndLines.Delete()
						currFuncEndLine = funcEndLines.Get()
					}

					message, isSecondArgumentString := getSecondArgumentName(expr)
					if !isSecondArgumentString {
						return false
					}

					if !checkCyrillic && hasCyrillic(message) {
						return false
					}

					funcName, ok := isFirstArgumentFuncCall(expr)
					if ok && !isWrapMsgSuitable(funcName, message) {
						printIncorrectWrap(fileName, position.Line)
					} else {
						errName, ok := getFirstArgumentName(expr)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:          position.Line,
								errName:       errName,
								message:       message,
								parentFunc:    currentFunc,
								nestingsLevel: nestingLevel,
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

		return strings.EqualFold(funcName, message)
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
		if !strings.EqualFold(messageSl[len(messageSl)-1-idx], funcNameSl[len(funcNameSl)-1-idx]) {
			return false
		}
	}

	return true
}

func checkWithUnderScore(funcName, message string) bool {
	return checkStandard(strings.ReplaceAll(funcName, "_", ""), strings.ReplaceAll(message, "_", ""))
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

func hasCyrillic(message string) bool {
	for _, r := range message {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}

	return false
}

func getWrapPackageName(node *ast.File) string {
	packageName := "errors"

	for _, imp := range node.Imports {
		importPath := imp.Path.Value[1 : len(imp.Path.Value)-1]

		if importPath == "github.com/pkg/errors" && imp.Name != nil {
			packageName = imp.Name.Name

			break
		}
	}

	return packageName
}

type Stack struct {
	sl []int
}

func (s *Stack) Add(n int) {
	s.sl = append(s.sl, n)
}

func (s *Stack) Get() int {
	if len(s.sl) > 0 {
		return s.sl[len(s.sl)-1]
	}

	return 0
}

// Delete does not remove last element
func (s *Stack) Delete() {
	if len(s.sl) <= 1 {
		return
	}

	s.sl = s.sl[:len(s.sl)-1]
}
