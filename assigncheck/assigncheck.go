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
		ast.Inspect(file, func(n ast.Node) bool {
			switch as := n.(type) {
			case *ast.AssignStmt:
				// new variable defined
				if as.Tok == token.DEFINE {
					return true
				}

				// there is one exception to this rule:
				// type assertions in the form x = x.(T),
				// as these do not change the value of x
				// at all.
				if _, ok := as.Rhs[0].(*ast.TypeAssertExpr); ok {
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
