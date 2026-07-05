package cwrapmsg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

func Handle() error {
	fileNames, err := getFileNames()
	if err != nil {
		return errors.Wrap(err, "getFileNames")
	}

	for _, fileName := range fileNames {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			return errors.Wrap(err, "parser.ParseFile")
		}

		wrapDataSl, errNamesMap := FindWrapCalls(node, fset, fileName)

		callDataMap := getFuncCalls(node, fset, maps.Keys(errNamesMap))

		findVariance(fileName, wrapDataSl, callDataMap)
	}

	return nil
}

func findVariance(fileName string, wrapDataSl []WrapData, callDataMap map[string][]CallData) {
	for _, wrapData := range wrapDataSl {
		var (
			needFuncName string
			index        int
		)

		for idx, callData := range callDataMap[wrapData.errName] {
			if callData.line >= wrapData.line {
				break
			}

			needFuncName = callData.funcName
			index = idx
		}

		if needFuncName == "" {
			continue
		}

		if !isWrapMsgSuitable(needFuncName, wrapData.message) {
			printIncorrectWrap(fileName, wrapData.line)
		}

		callDataMap[wrapData.errName] = slices.Delete(callDataMap[wrapData.errName], index, index+1)
	}
}

type CallData struct {
	line     int
	funcName string
}

func getFuncCalls(node *ast.File, fset *token.FileSet, errNames []string) map[string][]CallData {
	callDataMap := make(map[string][]CallData)

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		pos := fset.Position(callExpr.Pos())

		if parent := findParentAssignment(node, callExpr); parent != nil {
			if ident, ok := parent.(*ast.AssignStmt); ok {
				for _, lhs := range ident.Lhs {
					for _, errName := range errNames {
						if id, ok := lhs.(*ast.Ident); ok && id.Name == errName {
							funcName := getFuncName(callExpr.Fun)

							callDataMap[errName] = append(callDataMap[errName], CallData{
								line:     pos.Line,
								funcName: funcName,
							})
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
