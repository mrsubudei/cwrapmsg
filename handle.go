package cwrapmsg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

type Flags struct {
	SkipIgnoring bool
	OnlyUnstaged bool
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

		wrapDataSl, errNamesMap := FindWrapCalls(node, fset, fileName)

		callDataMap := getFuncCalls(node, fset, maps.Keys(errNamesMap))

		findVariance(fileName, wrapDataSl, callDataMap, ignoreDataMap)
	}

	return nil
}

func findVariance(
	fileName string,
	wrapDataSl []WrapData,
	callDataMap map[string][]CallData,
	ignoreDataMap map[string]IgnoreData,
) {
	for _, wrapData := range wrapDataSl {
		if ignoreData, hasFileName := ignoreDataMap[fileName]; hasFileName {
			if _, hasFuncName := ignoreData.funcNamesMap[wrapData.parentFunc]; hasFuncName {
				continue
			}
		}

		endLine := 0

		callDataSlice := callDataMap[wrapData.errName]
		sort.Slice(callDataSlice, func(i, j int) bool {
			return callDataSlice[i].endLine < callDataSlice[j].endLine
		})

		for _, callData := range callDataSlice {
			if callData.endLine >= wrapData.line {
				break
			}

			endLine = callData.endLine
		}

		callDataLineMap := getCallDataLineMap(callDataMap[wrapData.errName])
		callData, hasCalldata := callDataLineMap[endLine]

		if !hasCalldata || callData.funcName == "" {
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
	endLine    int
}

func getFuncCalls(node *ast.File, fset *token.FileSet, errNames []string) map[string][]CallData {
	callDataMap := make(map[string][]CallData)
	var currentFunc string

	ast.Inspect(node, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.FuncDecl:
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
		case *ast.CallExpr:
			endLine := fset.Position(expr.End())

			if parent := findParentAssignment(node, expr); parent != nil {
				if ident, ok := parent.(*ast.AssignStmt); ok {
					for _, lhs := range ident.Lhs {
						for _, errName := range errNames {
							if id, ok := lhs.(*ast.Ident); ok && id.Name == errName {
								funcName := getFuncName(expr.Fun)

								callDataMap[errName] = append(callDataMap[errName], CallData{
									funcName:   funcName,
									parentFunc: currentFunc,
									endLine:    endLine.Line,
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
	default:
		return fmt.Sprintf("%T", fun)
	}
}

func printIncorrectWrap(fileName string, line int) {
	fmt.Printf("%s %d\n", fileName, line)
}

func getCallDataLineMap(callDataSl []CallData) map[int]CallData {
	result := make(map[int]CallData, len(callDataSl))

	for _, v := range callDataSl {
		result[v.endLine] = v
	}

	return result
}
