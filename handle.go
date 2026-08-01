package cwrapmsg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

type Flags struct {
	SkipIgnoring  bool
	OnlyUnstaged  bool
	CheckCyrillic bool
}

func Handle(flags Flags) error {
	ignoreDataMap := GetIgnoreDataMap(flags.SkipIgnoring)

	unstagedFilesMap, err := getUnstagedFilesMap(flags.OnlyUnstaged)
	if err != nil {
		return errors.Wrap(err, "getUnstagedFilesMap")
	}

	fileNames, err := getFileNames()
	if err != nil {
		return errors.Wrap(err, "getFileNames")
	}

	for _, fileName := range fileNames {
		if _, ok := unstagedFilesMap[fileName]; !ok && flags.OnlyUnstaged {
			continue
		}

		chunks := getChunks(fileName)
		var skip bool

		for _, chunk := range chunks {
			if ignoreData, ok := ignoreDataMap[chunk]; ok && ignoreData.fullIgnore {
				skip = true
			}
		}

		if skip {
			continue
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			return errors.Wrap(err, "parser.ParseFile")
		}

		noLintsMap := findDisableLines(node, fset)

		wrapPackageName := getWrapPackageName(node)

		wrapDataSl, errNamesMap := FindWrapCalls(node, fset, fileName, wrapPackageName, flags.CheckCyrillic)

		callDataMap := getFuncCalls(node, fset, maps.Keys(errNamesMap))

		findVariance(fileName, wrapDataSl, callDataMap, ignoreDataMap, noLintsMap)
	}

	return nil
}

func findVariance(
	fileName string,
	wrapDataSl []WrapData,
	callDataMap map[string][]CallData,
	ignoreDataMap map[string]IgnoreData,
	noLintsMap map[int]struct{},
) {
	for _, wrapData := range wrapDataSl {
		if ignoreData, hasFileName := ignoreDataMap[fileName]; hasFileName {
			if _, hasFuncName := ignoreData.funcNamesMap[wrapData.parentFunc]; hasFuncName {
				continue
			}
		}

		if isWrapCallIgnored(wrapData.line, noLintsMap) {
			continue
		}

		errNameWithNesting := errNameAndNesting(wrapData.errName, wrapData.nestingsLevel)
		callDataSl := callDataMap[errNameWithNesting]

		callData, hasCallData := findCallData(callDataSl, wrapData.line)
		if !hasCallData {
			continue
		}

		if wrapData.parentFunc == callData.parentFunc && !isWrapMsgSuitable(callData.funcName, wrapData.message) {
			printIncorrectWrap(fileName, wrapData.line)
		}
	}
}

type CallData struct {
	funcName   string
	parentFunc string
	line       int
}

func getFuncCalls(node *ast.File, fset *token.FileSet, errNames []string) map[string][]CallData {
	callDataMap := make(map[string][]CallData)
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
				receiverType := ""
				switch t := expr.Recv.List[0].Type.(type) {
				case *ast.Ident:
					receiverType = t.Name
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						receiverType = ident.Name
					}
				}
				currentFunc = receiverType + "_" + expr.Name.Name
			} else {
				currentFunc = expr.Name.Name
			}
		case *ast.FuncLit:
			nestingLevel++
			position := fset.Position(expr.End())
			funcEndLines.Add(position.Line)
		case *ast.CallExpr:
			if parent := findParentAssignment(node, expr); parent != nil {
				if ident, ok := parent.(*ast.AssignStmt); ok {
					for _, lhs := range ident.Lhs {
						for _, errName := range errNames {
							if id, ok := lhs.(*ast.Ident); ok && id.Name == errName {
								position := fset.Position(expr.Pos())
								currFuncEndLine := funcEndLines.Get()
								for position.Line > currFuncEndLine {
									nestingLevel--
									funcEndLines.Delete()
									currFuncEndLine = funcEndLines.Get()
								}

								funcName := getFuncName(expr.Fun)
								errNameWithNesting := errNameAndNesting(errName, nestingLevel)

								callDataMap[errNameWithNesting] = append(callDataMap[errNameWithNesting], CallData{
									funcName:   funcName,
									parentFunc: currentFunc,
									line:       position.Line,
								})
							}
						}
					}
				}
			}
		}

		return true
	})

	return callDataMap
}

func findParentAssignment(node ast.Node, callExpr *ast.CallExpr) ast.Node {
	var parent ast.Node
	ast.Inspect(node, func(n ast.Node) bool {
		if n == callExpr {
			return false
		}
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, rhs := range assign.Rhs {
				if rhs == callExpr {
					parent = n

					return false
				}
			}
		}

		return true
	})

	return parent
}

func getFuncName(fun ast.Expr) string {
	switch expr := fun.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getFuncName(expr.X), expr.Sel.Name)
	case *ast.IndexExpr:
		return getFuncName(expr.X)
	case *ast.IndexListExpr:
		return getFuncName(expr.X)
	default:
		return fmt.Sprintf("%T", fun)
	}
}

func printIncorrectWrap(fileName string, line int) {
	fmt.Printf("%s %d\n", fileName, line)
}

func findCallData(callDataSl []CallData, wrapLine int) (CallData, bool) {
	result := CallData{}
	found := false

	for _, callData := range callDataSl {
		if callData.line < wrapLine {
			result = callData
			found = true
		} else {
			break
		}
	}

	return result, found
}

func errNameAndNesting(errName string, nestingLevel int) string {
	return fmt.Sprintf("%s-%d", errName, nestingLevel)
}

func findDisableLines(file *ast.File, fset *token.FileSet) map[int]struct{} {
	result := make(map[int]struct{})

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if strings.HasPrefix(comment.Text, "//nolint:") {
				linters := strings.TrimPrefix(comment.Text, "//nolint:")
				linters = strings.TrimSpace(linters)

				for _, linter := range strings.Split(linters, ",") {
					if strings.TrimSpace(linter) == "cwrap" {
						pos := fset.Position(comment.Pos())

						result[pos.Line] = struct{}{}
					}
				}
			}
		}
	}

	return result
}

func isWrapCallIgnored(callLine int, noLintMap map[int]struct{}) bool {
	if _, ok := noLintMap[callLine]; ok {
		return true
	}

	if _, ok := noLintMap[callLine-1]; ok {
		return true
	}

	return false
}
