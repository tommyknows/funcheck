package assigncheck

import (
	"go/ast"
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
		// function assignments, if that function was just recently
		// declared, should be allowed. Anonymous functions cannot be
		// called recursively if they are not in scope yet. This means
		// that to call an anonymous function, the following pattern
		// is always needed:
		//   var x func(int) string
		//   x = func(int) string { ... x(int) }
		// To ignore that, whenever a "var x func" is encountered, we
		// save that position until the next node.
		var lastFuncDecl token.Pos

		ast.Inspect(file, func(n ast.Node) bool {
			switch as := n.(type) {
			case *ast.DeclStmt:
				lastFuncDecl = functionPos(as)
				return false // important to return, as we'd reset the position if not

			case *ast.AssignStmt:
				for _, i := range identReassigned(as, lastFuncDecl) {
					pass.Reportf(as.Pos(), "re-assignment of %s", i)
				}

			case *ast.IncDecStmt:
				pass.Reportf(as.Pos(), "inline re-assignment of %s", as.X)
			}

			lastFuncDecl = token.NoPos
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
func identReassigned(as *ast.AssignStmt, lastFuncPos token.Pos) []*ast.Ident {
	type pos interface {
		Pos() token.Pos
	}

	var reassigned []*ast.Ident

	var expectedAssignPos token.Pos
	for i, expr := range as.Lhs {
		ident := expr.(*ast.Ident) // Lhs always is an "IdentifierList"

		// skip blank identifiers
		if ident.Obj == nil {
			continue
		}

		// make sure the declaration has a Pos func and get it
		declPos := ident.Obj.Decl.(pos).Pos()

		// if we got a function position and the corresponding
		// Rhs expression is a function literal, check that the
		// positions match (=same declaration)
		if lastFuncPos != token.NoPos && len(as.Rhs) > i {
			if _, ok := as.Rhs[i].(*ast.FuncLit); ok {
				// the function is either declared right here or on the last
				// position that we got from the callee.
				if declPos != lastFuncPos && declPos != ident.Pos() {
					reassigned = append(reassigned, ident)
				}
				continue
			}
		}

		// we expect all assignments to be at the same position
		// as the first identifier.
		// paired with the below if condition, this bascially
		// ensures that the first identifier's assignment is
		// at the same position as the identifier itself.
		if expectedAssignPos == token.NoPos {
			expectedAssignPos = ident.Pos()
		}

		if declPos != expectedAssignPos {
			reassigned = append(reassigned, ident)
		}
	}

	return reassigned
}

// functionPos returns the position of the function
// declaration, if the DeclStmt has a function declaration
// at all. If not, token.NoPos is returned.
// At most, one position (the position of the last function
// declaration) is returned
func functionPos(as *ast.DeclStmt) token.Pos {
	decl, ok := as.Decl.(*ast.GenDecl)
	if !ok {
		return token.NoPos
	}

	var pos token.Pos

	// iterate over all variable specs to fetch
	// the last function declaration. Skip all declarations
	// that are not function literals.
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

		// there may not be more than one function
		// declaration mapped to a single type,
		// so we just return the first one.
		pos = val.Names[0].Pos()
	}
	return pos
}
