package e4qa

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/reusee/qa"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func (_ Def) CheckHandleAndCheckUsage(
	pkgs []*packages.Package,
	checkFunc CheckFuncObject,
	handleFunc HandleFuncObject,
) qa.CheckFunc {
	return func() {

		// check
		packages.Visit(pkgs, func(pkg *packages.Package) bool {

			// get Check and Handle aliases
			checkObjects := FindAlias(pkg, checkFunc.Object)
			handleObjects := FindAlias(pkg, handleFunc.Object)

			// error type
			tv, err := types.Eval(pkg.Fset, pkg.Types, 0, "error(nil)")
			ce(err)
			errorType := tv.Type

			// check usages
			notHandled := make(map[token.Pos][]token.Pos)
			for _, check := range checkObjects {
				for checkIdent, obj := range pkg.TypesInfo.Uses {
					if obj != check {
						continue
					}
					for _, file := range pkg.Syntax {
						path, exact := astutil.PathEnclosingInterval(file, checkIdent.Pos(), checkIdent.Pos())
						if !exact {
							continue
						}
						for _, node := range path {

							var body *ast.BlockStmt
							var signature *ast.FuncType

							if fnLit, ok := node.(*ast.FuncLit); ok {
								// function literal
								fnSig := pkg.TypesInfo.Types[fnLit].Type.(*types.Signature)
								rets := fnSig.Results()
								if rets.Len() < 1 {
									continue
								}
								if types.Identical(
									rets.At(rets.Len()-1).Type(),
									errorType,
								) {
									body = fnLit.Body
									signature = fnLit.Type
								}

							} else if fnDecl, ok := node.(*ast.FuncDecl); ok {
								// function decl
								fnSig := pkg.TypesInfo.Defs[fnDecl.Name].Type().(*types.Signature)
								rets := fnSig.Results()
								if rets.Len() < 1 {
									continue
								}
								if types.Identical(
									rets.At(rets.Len()-1).Type(),
									errorType,
								) {
									body = fnDecl.Body
									signature = fnDecl.Type
								}
							}

							if body != nil {
								checkOK := false
								// find handle call before check call
								for _, stmt := range body.List {
									if stmt.Pos() > checkIdent.End() {
										// stmt after check
										break
									}
									deferStmt, ok := stmt.(*ast.DeferStmt)
									if !ok {
										// not defer
										continue
									}
									callIdent, ok := deferStmt.Call.Fun.(*ast.Ident)
									if !ok {
										// not call by identifier
										continue
									}
									callObj := pkg.TypesInfo.Uses[callIdent]
									isHandleObject := false
									for _, obj := range handleObjects {
										if callObj == obj {
											isHandleObject = true
											break
										}
									}
									if !isHandleObject {
										// not handle object
										continue
									}
									// check target argument of handle call
									target := deferStmt.Call.Args[0]
									targetExpr, ok := target.(*ast.UnaryExpr)
									if !ok {
										pt("expecting unary expression: %s\n", pkg.Fset.Position(target.Pos()).String())
										return false
									}
									errIdent, ok := targetExpr.X.(*ast.Ident)
									if !ok {
										pt("expecting error identifier: %s\n", pkg.Fset.Position(target.Pos()).String())
										return false
									}
									errObj := pkg.TypesInfo.Uses[errIdent]
									// find def of error object
									for defIdent, defObj := range pkg.TypesInfo.Defs {
										if defObj != errObj {
											continue
										}
										// must define inside signature
										if !(defIdent.Pos() > signature.Results.Pos() && defIdent.End() < signature.Results.End()) {
											pt("should pass error defined at %s\n",
												pkg.Fset.Position(signature.Results.Pos()))
											return false
										}
										checkOK = true
									}

								}

								if !checkOK {
									body := body.Pos()
									checkPos := checkIdent.Pos()
									notHandled[body] = append(
										notHandled[body],
										checkPos,
									)
								}

								break
							}

						}
					}
				}
			}

			for bodyPos, checkPoses := range notHandled {
				pt("not handled %s\n",
					pkg.Fset.Position(bodyPos).String())
				for _, pos := range checkPoses {
					pt("\t%s\n", pkg.Fset.Position(pos).String())
				}
			}

			return true
		}, nil)
	}
}
