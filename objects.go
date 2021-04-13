package e4qa

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type CheckFuncObject struct {
	types.Object
}

type HandleFuncObject struct {
	types.Object
}

type TestingFatalFuncObject struct {
	types.Object
}

func (_ Def) Objects(
	pkgs []*packages.Package,
) (
	check CheckFuncObject,
	handle HandleFuncObject,
	testingFatal TestingFatalFuncObject,
) {

	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		if pkg.PkgPath == "github.com/reusee/e4" {
			check.Object = pkg.Types.Scope().Lookup("Check")
			handle.Object = pkg.Types.Scope().Lookup("Handle")
			testingFatal.Object = pkg.Types.Scope().Lookup("TestingFatal")
		}
		return true
	}, nil)

	return
}

func FindAlias(pkg *packages.Package, target types.Object) []types.Object {
	objs := []types.Object{
		target,
	}
	for ident, obj := range pkg.TypesInfo.Uses {
		if obj != target {
			continue
		}
		objs = append(objs, obj)
		for _, file := range pkg.Syntax {
			path, exact := astutil.PathEnclosingInterval(
				file, ident.Pos(), ident.End())
			if !exact {
				continue
			}
			for _, node := range path {
				valueSpec, ok := node.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, value := range valueSpec.Values {
					selExpr, ok := value.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					if selExpr.Sel != ident {
						continue
					}
					id := valueSpec.Names[i]
					obj := pkg.TypesInfo.Defs[id]
					objs = append(objs, obj)
				}
			}
		}
	}
	return objs
}
