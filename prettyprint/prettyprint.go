package prettyprint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "prettyprint",
	Doc:  "prints positions",
	Run:  run,
}

type null struct{}

func checkDecl(as *ast.DeclStmt, fset *token.FileSet) {
	fmt.Printf("Declaration %q: %v\n", render(fset, as), as.Pos())

	check := func(_ null, spec ast.Spec) (n null) {
		val, ok := spec.(*ast.ValueSpec)
		if !ok {
			return
		}
		if val.Values != nil {
			return
		}
		if _, ok := val.Type.(*ast.FuncType); !ok {
			return
		}
		fmt.Printf("\tIdent %q: %v\n", render(fset, val), val.Names[0].Pos())
		return
	}

	if decl, ok := as.Decl.(*ast.GenDecl); ok {
		_ = foldl(check, null{}, decl.Specs)
	}
}

func checkAssign(as *ast.AssignStmt, fset *token.FileSet) {
	fmt.Printf("Assignment %q: %v\n", render(fset, as), as.Pos())

	check := func(_ null, expr ast.Expr) (n null) {
		ident, ok := expr.(*ast.Ident) // Lhs always is an "IdentifierList"
		if !ok {
			return
		}

		fmt.Printf("\tIdent %q: %v\n", ident.String(), ident.Pos())

		switch {
		case ident.Name == "_":
			fmt.Printf("\t\tBlank Identifier!\n")
		case ident.Obj == nil:
			fmt.Printf("\t\tDecl is not in the same file!\n")
		default:
			// make sure the declaration has a Pos func and get it
			declPos := ident.Obj.Decl.(ast.Node).Pos()
			fmt.Printf("\t\tDecl %q: %v\n", render(fset, ident.Obj.Decl), declPos)
		}

		return
	}
	_ = foldl(check, null{}, as.Lhs)
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := func(_ null, file *ast.File) (n null) {
		ast.Inspect(file, func(n ast.Node) bool {
			switch as := n.(type) {
			case *ast.DeclStmt:
				checkDecl(as, pass.Fset)
			case *ast.AssignStmt:
				checkAssign(as, pass.Fset)
			}
			return true
		})
		return
	}
	_ = foldl(inspect, null{}, pass.Files)

	return nil, nil
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
