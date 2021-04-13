package main

import (
	"github.com/reusee/dscope"
	"github.com/reusee/e4qa"
	"github.com/reusee/qa"
)

func main() {

	defs := dscope.Methods(new(qa.Def))
	defs = append(defs, dscope.Methods(new(e4qa.Def))...)

	scope := dscope.New(defs...)

	var check qa.CheckFunc
	scope.Assign(&check)
	check()

}
