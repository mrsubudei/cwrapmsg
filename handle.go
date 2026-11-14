package cwrapmsg

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/analysis"
)

type Config struct {
	IgnoreCyrillic  bool
	CaseInsensitive bool
}

func handle(
	pass *analysis.Pass,
	file *ast.File,
	cfg Config,
	wrapPackageName, fmtPackageName string,
) {
	wrapDataSl, errNamesMap := findWrapCalls(pass, file, pass.Fset, cfg, wrapPackageName, fmtPackageName)

	callDataMap := getFuncCalls(file, pass.Fset, maps.Keys(errNamesMap))

	findVariance(pass, cfg, wrapDataSl, callDataMap)
}

func findVariance(
	pass *analysis.Pass,
	cfg Config,
	wrapDataSl []WrapData,
	callDataMap map[string][]CallData,
) {
	for _, wrapData := range wrapDataSl {
		errNameWithNesting := errNameAndNesting(wrapData.errName, wrapData.nestingsLevel)
		callDataSl := callDataMap[errNameWithNesting]

		callData, hasCallData := findCallData(callDataSl, wrapData.line)
		if !hasCallData {
			continue
		}

		if wrapData.parentFunc == callData.parentFunc && !isWrapMsgSuitable(cfg, callData.funcName, wrapData.message) {
			pass.Report(getDiagReport(wrapData.node, wrapData.isWrapf, wrapData.isErrorf))
		}
	}
}

func getDiagReport(node ast.Node, isWrapf, isErrorf bool) analysis.Diagnostic {
	prefix := "errors.Wrap"

	if isWrapf {
		prefix += "f"
	}
	if isErrorf {
		prefix = "fmt.Errorf"
	}

	return analysis.Diagnostic{
		Pos:     node.Pos(),
		End:     node.End(),
		Message: fmt.Sprintf("message in %s does not match function or method name", prefix),
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
			if parent := findParentAssignment(node, expr); parent != nil {
				if lhsList, ok := getAssignedNames(parent); ok {
					for _, lhs := range lhsList {
						for _, errName := range errNames {
							if id, ok := lhs.(*ast.Ident); ok && id.Name == errName {
								position := fset.Position(expr.Pos())
								currFuncEndLine := funcEndLines.Get()
								for position.Line > currFuncEndLine && nestingLevel > 0 {
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

		if declStmt, ok := n.(*ast.DeclStmt); ok {
			if genDecl, ok := declStmt.Decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, value := range valueSpec.Values {
							if value == callExpr {
								parent = valueSpec

								return false
							}
						}
					}
				}
			}
		}

		return true
	})

	return parent
}

func getAssignedNames(parent ast.Node) ([]ast.Expr, bool) {
	switch stmt := parent.(type) {
	case *ast.AssignStmt:
		return stmt.Lhs, true
	case *ast.ValueSpec:
		names := make([]ast.Expr, 0, len(stmt.Names))
		for _, name := range stmt.Names {
			names = append(names, name)
		}

		return names, true
	}

	return nil, false
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
