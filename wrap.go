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
	column        int
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
	fmtPackageName string,
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

					message, isSecondArgumentString := getArgumentValue(expr, 1)
					if !isSecondArgumentString {
						return false
					}

					if !checkCyrillic && hasCyrillic(message) {
						return false
					}

					funcName, ok := isArgumentFuncCall(expr, 0)
					if ok && !isWrapMsgSuitable(funcName, message) {
						printIncorrectWrap(fileName, position.Line, position.Column)
					} else {
						errName, ok := getArgumentName(expr, 0)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:          position.Line,
								column:        position.Column,
								errName:       errName,
								message:       message,
								parentFunc:    currentFunc,
								nestingsLevel: nestingLevel,
							})
						}
					}
				}
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == fmtPackageName &&
					(fun.Sel.Name == "Errorf") && len(expr.Args) >= 2 {

					position := fset.Position(expr.Pos())
					currFuncEndLine := funcEndLines.Get()
					for position.Line > currFuncEndLine {
						nestingLevel--
						funcEndLines.Delete()
						currFuncEndLine = funcEndLines.Get()
					}

					message, ok := getArgumentValue(expr, 0)
					if !ok || !strings.Contains(message, "%w") {
						return false
					}

					wIdx, ok := getWIndex(message)
					if !ok || len(expr.Args) <= wIdx+1 {
						return false
					}

					if !checkCyrillic && hasCyrillic(message) {
						return false
					}

					funcName, ok := isArgumentFuncCall(expr, wIdx+1)
					if ok && !isWrapMsgSuitable(funcName, message) {
						printIncorrectWrap(fileName, position.Line, position.Column)
					} else {
						errName, ok := getArgumentName(expr, wIdx+1)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:          position.Line,
								column:        position.Column,
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

		return strings.EqualFold(funcName, cutMessage(message))
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

func getArgumentName(callExpr *ast.CallExpr, idx int) (string, bool) {
	if firstArg, ok := callExpr.Args[idx].(*ast.Ident); ok {
		return firstArg.Name, true
	}

	return "", false
}

func getArgumentValue(callExpr *ast.CallExpr, idx int) (string, bool) {
	if strLit, ok := callExpr.Args[idx].(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		return strings.Trim(strLit.Value, "\\\""), true
	}

	return "", false
}

func isArgumentFuncCall(callExpr *ast.CallExpr, idx int) (string, bool) {
	if callExp, isFuncCall := callExpr.Args[idx].(*ast.CallExpr); isFuncCall {
		return getFuncName(callExp.Fun), true
	}

	return "", false
}

func getWIndex(str string) (int, bool) {
	count := 0
	letterIdx := -1
	wIdx := 0

	for i := 0; i < len(str)-1; i++ {
		if str[i] == '%' {
			if unicode.IsLetter(rune(str[i+1])) {
				letterIdx++

				if str[i+1] == 'w' {
					count++
					wIdx = letterIdx
				}
			}
		}
	}

	return wIdx, count == 1
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

func getFmtPackageName(node *ast.File) string {
	packageName := "fmt"

	for _, imp := range node.Imports {
		importPath := imp.Path.Value[1 : len(imp.Path.Value)-1]

		if importPath == "fmt" && imp.Name != nil {
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
