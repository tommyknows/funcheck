package assigncheck

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "assigncheck",
	Doc:  "reports reassignments",
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
			case *ast.ForStmt:
				// exception for all for loops that do not have a post-condition.
				// a loop without a post-condition technically only makes sense if
				// the variable could be mutated within the loop, which they can't
				// either. But we're being correct here :-)
				if as.Post != nil {
					pass.Reportf(as.Pos(), "reassignment (for loop) in %q", renderFor(pass.Fset, as))
				}
				return true
			case *ast.RangeStmt:
				pass.Reportf(as.Pos(), "internal reassignment (for loop) in %q", renderRange(pass.Fset, as))
				return true

			case *ast.IncDecStmt:
				pass.Reportf(as.Pos(), "inline reassignment of %s", render(pass.Fset, as.X))

			case *ast.DeclStmt:
				lastFuncDecl = functionPos(as)
				return false // important to return, as we'd reset the position if not

			case *ast.AssignStmt:
				for _, i := range exprReassigned(as, lastFuncDecl) {
					pass.Reportf(as.Pos(), "reassignment of %s", render(pass.Fset, i))
				}
			}

			lastFuncDecl = token.NoPos
			return true
		})
	}

	return nil, nil
}

func renderFor(fset *token.FileSet, as *ast.ForStmt) string {
	s := "for "
	if as.Init == nil && as.Cond == nil && as.Post == nil {
		return s + "{ ... }"
	}

	if as.Init == nil && as.Cond != nil && as.Post == nil {
		return s + render(fset, as.Cond) + " { ... }"
	}

	if as.Init != nil {
		s += render(fset, as.Init)
	}
	s += "; "
	if as.Cond != nil {
		s += render(fset, as.Cond)
	}
	s += "; "
	if as.Post != nil {
		s += render(fset, as.Post)
	}

	return s + " { ... }"
}

func renderRange(fset *token.FileSet, as *ast.RangeStmt) string {
	s := "for "
	switch {
	case as.Key == nil && as.Value == nil:
		// nothing
	case as.Value == nil:
		s += render(fset, as.Key) + " := "
	case as.Key == nil:
		s += "_, " + render(fset, as.Value) + " := "
	default:
		s += render(fset, as.Key) + ", " + render(fset, as.Value) + " := "
	}
	s += "range " + render(fset, as.X) + " { ... }"
	return s
}

const blankIdent = "_"

// isTestdeps returns true if the expression is the
// testdeps.ImportPath selector. This selector is generated
// as testing code and is not anything the programmer could
// influence.
func isTestdeps(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	i, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return i.Name == "testdeps" && sel.Sel.Name == "ImportPath"
}

// exprReassigned returns all expressions in an assignment
// that are being reassigned. This is done by checking that the
// assignment of all identifiers is at the position of the first
// identifier. If it is not an identifier, it must be a reassignment.
// There are two exceptions to this rule:
// - Blank identifiers are ignored
// - Functions may be redeclared if the assignment position is the lastFuncPos
func exprReassigned(as *ast.AssignStmt, lastFuncPos token.Pos) (reassigned []ast.Expr) {
	type pos interface {
		Pos() token.Pos
	}

	var expectedAssignPos token.Pos

	for i, expr := range as.Lhs {
		ident, ok := expr.(*ast.Ident)
		if !ok { // if it's not an identifier, it is always reassigned.
			if !isTestdeps(expr) { // ignore testdeps.ImportPath reassignment
				reassigned = append(reassigned, expr)
			}
			continue
		}

		// we expect all assignments to be at the same position
		// as the first identifier.
		if expectedAssignPos == token.NoPos {
			expectedAssignPos = ident.Pos()
		}

		// skip blank identifiers
		if ident.Name == blankIdent {
			continue
		}

		// no object probably means that the variable has been declared
		// in a separate file, making this a reassignment.
		if ident.Obj == nil {
			reassigned = append(reassigned, ident)
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
				if declPos != lastFuncPos && declPos != expectedAssignPos {
					reassigned = append(reassigned, ident)
				}
				continue
			}
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

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
