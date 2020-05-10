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
				fmt.Printf("Declaration %q: %v\n", render(pass.Fset, as), as.Pos())
				decl, ok := as.Decl.(*ast.GenDecl)
				if !ok {
					break
				}

				for i := range decl.Specs {
					val, ok := decl.Specs[i].(*ast.ValueSpec)
					if !ok {
						continue
					}

					if val.Values != nil {
						continue
					}

					_, ok = val.Type.(*ast.FuncType)
					if !ok {
						continue
					}

					fmt.Printf("\tIdent %q: %v\n", render(pass.Fset, val), val.Names[0].Pos())
				}
			case *ast.AssignStmt:
				type pos interface {
					Pos() token.Pos
				}

				fmt.Printf("Assignment %q: %v\n", render(pass.Fset, as), as.Pos())

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
					fmt.Printf("\t\tDecl %q: %v\n", render(pass.Fset, ident.Obj.Decl), declPos)
				}
			}
			return true
		})
	}

	return nil, nil
}

// identReassigned returns all identifiers in an assignment
// that are being reassigned. This is done by checking that the
// assignment of all identifiers is at the position of the first
// identifier.
// There are two exceptions to this rule:
// - Blank identifiers are ignored
// - Functions may be redeclared if the assignment position is
//   the lastFuncPos
//func identReassigned(as *ast.AssignStmt) {
//type pos interface {
//Pos() token.Pos
//}

//fmt.Printf("Assignment %q Position: %v\n", render(pass.Fset, as), as.Pos())

//for _, expr := range as.Lhs {
//ident := expr.(*ast.Ident) // Lhs always is an "IdentifierList"

//fmt.Printf("  Ident %q Position: %v\n", ident.String(), ident.Pos())

//// skip blank identifiers
//if ident.Obj == nil {
//continue
//}

//// make sure the declaration has a Pos func and get it
//declPos := ident.Obj.Decl.(pos).Pos()
//fmt.Printf("  Ident Decl Position: %v\n", declPos)
//}
//}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
