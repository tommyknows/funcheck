/*
The prettyprint package visualises a Go AST, but only assignment
nodes. In these nodes, it prints their position, the identifier
and the identifier's declaration plus position.

It is used to visualise how assigncheck is checking for re-assignments.

To see it for yourself, run `go test -v .` in this directory.

	$> go test -v .
	=== RUN   TestRun
	Assignment "x, y := 1, 2": 2958101
	        Ident "x": 2958101
	                Decl "x, y := 1, 2": 2958101
	        Ident "y": 2958104
	                Decl "x, y := 1, 2": 2958101
	Assignment "y = 3": 2958115
	        Ident "y": 2958115
	                Decl "x, y := 1, 2": 2958101
	--- PASS: TestRun (0.93s)
	PASS
	ok      github.com/tommyknows/funcheck/prettyprint      1.088s
*/
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

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch as := n.(type) {
			case *ast.DeclStmt:
				checkDecl(as, pass.Fset)
			case *ast.AssignStmt:
				checkAssign(as, pass.Fset)
			}
			return true
		})
	}

	return nil, nil
}

func checkDecl(as *ast.DeclStmt, fset *token.FileSet) {
	fmt.Printf("Declaration %q: %v\n", render(fset, as), as.Pos())
	decl, ok := as.Decl.(*ast.GenDecl)
	if !ok {
		return
	}

	for i := range decl.Specs {
		val, ok := decl.Specs[i].(*ast.ValueSpec)
		if !ok {
			continue
		}

		if val.Values != nil {
			continue
		}

		if _, ok := val.Type.(*ast.FuncType); !ok {
			continue
		}

		fmt.Printf("\tIdent %q: %v\n", render(fset, val), val.Names[0].Pos())
	}
}

func checkAssign(as *ast.AssignStmt, fset *token.FileSet) {
	type pos interface {
		Pos() token.Pos
	}

	fmt.Printf("Assignment %q: %v\n", render(fset, as), as.Pos())

	for _, expr := range as.Lhs {
		ident := expr.(*ast.Ident) // Lhs always is an "IdentifierList"

		fmt.Printf("\tIdent %q: %v\n", ident.String(), ident.Pos())

		// skip blank identifiers
		if ident.Name == "_" {
			fmt.Printf("\t\tBlank Identifier!\n")
			continue
		}

		if ident.Obj == nil {
			fmt.Printf("\t\tDecl is not in the same file!\n")
			continue
		}

		// make sure the declaration has a Pos func and get it
		declPos := ident.Obj.Decl.(pos).Pos()
		fmt.Printf("\t\tDecl %q: %v\n", render(fset, ident.Obj.Decl), declPos)
	}
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
