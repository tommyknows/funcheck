package assigncheck

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "assigncheck",
	Doc:  "reports re-assignments",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		var lastNodeFuncDeclName string
		ast.Inspect(file, func(n ast.Node) bool {
			switch as := n.(type) {
			// check if we found the declaration of a function,
			// and store its name if so. This is needed to have
			// recursive anonymous function, see comment below.
			case *ast.DeclStmt:
				decl, ok := as.Decl.(*ast.GenDecl)
				if !ok {
					return true
				}

				val, ok := decl.Specs[0].(*ast.ValueSpec)
				if !ok {
					return true
				}
				if val.Values != nil {
					return true
				}

				_, ok = val.Type.(*ast.FuncType)
				if !ok {
					return true
				}

				lastNodeFuncDeclName = val.Names[0].Name
				return false
			case *ast.AssignStmt:

				// new variable defined
				if as.Tok == token.DEFINE {
					return true
				}

				// there are two exceptions to this rule:
				// ONE: type assertions in the form x = x.(T),
				//      as these do not change the value of x.
				// TWO: function assignments if it was just
				//      recently declared. Anonymous functions
				//      cannot be called recursively if they
				//      are not in scope yet. This means that
				//      to call an anonymous function, the
				//      following pattern is always needed:
				//        var x func(int) string
				//        x = func(int) string { ... x(int) }
				//      To ignore that, whenever a "var x func"
				//      is seen, we save that identifier until
				//      the next node.
				switch as.Rhs[0].(type) {
				case *ast.TypeAssertExpr:
					return true
				case *ast.FuncLit:
					expr := as.Lhs[0].(*ast.Ident)
					if expr.Name == lastNodeFuncDeclName {
						lastNodeFuncDeclName = ""
						return true
					}
				}

				// ignore blank identifiers
				if i, ok := as.Lhs[0].(*ast.Ident); ok && i.Obj == nil {
					return true
				}

				pass.Reportf(as.Pos(), "re-assignment of %s",
					renderExpressions(pass.Fset, as.Lhs),
				)

			case *ast.IncDecStmt:
				pass.Reportf(as.Pos(), "inline re-assignment of %s",
					renderIncDec(pass.Fset, as.X),
				)
			}

			lastNodeFuncDeclName = "" // reset the name, we just passed a node.
			return true
		})
	}

	return nil, nil
}

// renderExpressions returns the pretty-print of multiple expressions.
// For example, the left-handside of an assigment
//   x, y := 0, 0
// is returned as "x, y"
func renderExpressions(fset *token.FileSet, x []ast.Expr) string {
	var buf bytes.Buffer
	for i, e := range x {
		if err := printer.Fprint(&buf, fset, e); err != nil {
			panic(err)
		}
		if !(len(x)-1 == i) {
			fmt.Fprintf(&buf, ", ")
		}
	}
	return buf.String()
}

func renderIncDec(fset *token.FileSet, x ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
