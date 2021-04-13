package e4qa

import (
	"go/ast"

	"github.com/reusee/qa"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func (_ Def) CehckTestingFatalWithoutHandle(
	pkgs []*packages.Package,
	testingFatalFunc TestingFatalFuncObject,
	handleFunc HandleFuncObject,
) qa.CheckFunc {
	return func() {

		packages.Visit(pkgs, func(pkg *packages.Package) bool {

			handleObjects := FindAlias(pkg, handleFunc.Object)

			for ident, obj := range pkg.TypesInfo.Uses {
				if obj != testingFatalFunc.Object {
					continue
				}

				for _, file := range pkg.Syntax {
					path, exact := astutil.PathEnclosingInterval(
						file, ident.Pos(), ident.End())
					if !exact {
						continue
					}

					inDefer := false
					for _, node := range path {
						if _, ok := node.(*ast.DeferStmt); ok {
							inDefer = true
							break
						}
					}
					if !inDefer {
						continue
					}

					inHandle := false
					for _, node := range path {
						call, ok := node.(*ast.CallExpr)
						if !ok {
							continue
						}
						var fnIdent *ast.Ident
						if id, ok := call.Fun.(*ast.Ident); ok {
							fnIdent = id
						} else if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							fnIdent = sel.Sel
						}
						obj := pkg.TypesInfo.Uses[fnIdent]
						for _, alias := range handleObjects {
							if alias == obj {
								inHandle = true
							}
						}
					}

					if !inHandle {
						pt("TestingFatal should be used inside Handle: %s\n",
							pkg.Fset.Position(ident.Pos()).String())
					}

				}
			}

			return true
		}, nil)

	}
}
