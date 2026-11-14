package cwrapmsg

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

type WrapData struct {
	errName       string
	message       string
	parentFunc    string
	nestingsLevel int
	line          int
	node          ast.Node
	isErrorf      bool
	isWrapf       bool
}

func findWrapCalls(
	pass *analysis.Pass,
	file *ast.File,
	fset *token.FileSet,
	cfg Config,
	wrapPackageName string,
	fmtPackageName string,
) ([]WrapData, map[string]struct{}) {
	wrapDataSl := []WrapData{}
	errorNamesMap := make(map[string]struct{})
	var currentFunc string
	nestingLevel := 0
	funcEndLines := Stack{}

	ast.Inspect(file, func(n ast.Node) bool {
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
			position := fset.Position(expr.End())
			currentFuncEndLine := funcEndLines.Get()

			for position.Line > currentFuncEndLine && nestingLevel > 0 {
				nestingLevel--
				funcEndLines.Delete()
				currentFuncEndLine = funcEndLines.Get()
			}

			nestingLevel++
			funcEndLines.Add(position.Line)
		case *ast.CallExpr:
			if fun, ok := expr.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == wrapPackageName &&
					(fun.Sel.Name == "Wrap" || fun.Sel.Name == "Wrapf") && len(expr.Args) >= 2 {

					position := fset.Position(expr.Pos())
					currFuncEndLine := funcEndLines.Get()
					for position.Line > currFuncEndLine && nestingLevel > 0 {
						nestingLevel--
						funcEndLines.Delete()
						currFuncEndLine = funcEndLines.Get()
					}

					message, isSecondArgumentString := getArgumentValue(expr, 1)
					if !isSecondArgumentString {
						return false
					}

					if cfg.IgnoreCyrillic && hasCyrillic(message) {
						return false
					}

					isWrapf := fun.Sel.Name == "Wrapf"

					funcName, ok := isArgumentFuncCall(expr, 0)
					if ok && !isWrapMsgSuitable(cfg, funcName, message) {
						pass.Report(getDiagReport(expr, isWrapf, false))
					} else {
						errName, ok := getArgumentName(expr, 0)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:          position.Line,
								errName:       errName,
								message:       message,
								parentFunc:    currentFunc,
								nestingsLevel: nestingLevel,
								node:          expr,
								isWrapf:       isWrapf,
							})
						}
					}
				}
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == fmtPackageName &&
					(fun.Sel.Name == "Errorf") && len(expr.Args) >= 2 {

					position := fset.Position(expr.Pos())
					currFuncEndLine := funcEndLines.Get()
					for position.Line > currFuncEndLine && nestingLevel > 0 {
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

					if cfg.IgnoreCyrillic && hasCyrillic(message) {
						return false
					}

					funcName, ok := isArgumentFuncCall(expr, wIdx+1)
					if ok && !isWrapMsgSuitable(cfg, funcName, message) {
						pass.Report(getDiagReport(expr, false, true))
					} else {
						errName, ok := getArgumentName(expr, wIdx+1)
						if ok {
							errorNamesMap[errName] = struct{}{}
							wrapDataSl = append(wrapDataSl, WrapData{
								line:          position.Line,
								errName:       errName,
								message:       message,
								parentFunc:    currentFunc,
								nestingsLevel: nestingLevel,
								node:          expr,
								isErrorf:      true,
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

func isWrapMsgSuitable(cfg Config, funcName, message string) bool {
	if strings.Contains(funcName, "*ast.CallExpr") {
		funcName = getLastWord(funcName)
		message = getLastWord(message)

		return isEqual(cfg, funcName, cutMessage(message))
	}

	if !checkStandard(cfg, funcName, message) && !checkWithUnderScore(cfg, funcName, message) {
		return false
	}

	return true
}

func isEqual(cfg Config, funcName, message string) bool {
	if cfg.CaseInsensitive {
		return strings.EqualFold(funcName, message)
	}

	return funcName == message
}

func checkStandard(cfg Config, funcName, message string) bool {
	funcNameSl := strings.Split(funcName, ".")
	messageSl := strings.Split(cutMessage(message), ".")

	if len(messageSl) > len(funcNameSl) {
		return false
	}

	for idx := 0; idx < len(messageSl); idx++ {
		if !isEqual(cfg, funcNameSl[len(funcNameSl)-1-idx], messageSl[len(messageSl)-1-idx]) {
			return false
		}
	}

	return true
}

func checkWithUnderScore(cfg Config, funcName, message string) bool {
	return checkStandard(cfg, strings.ReplaceAll(funcName, "_", ""), strings.ReplaceAll(message, "_", ""))
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

func getPackageNameOrAlias(file *ast.File, defaultName, packageImport string) string {
	for _, imp := range file.Imports {
		importPath := imp.Path.Value[1 : len(imp.Path.Value)-1]

		if importPath == packageImport && imp.Name != nil {
			defaultName = imp.Name.Name

			break
		}
	}

	return defaultName
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
