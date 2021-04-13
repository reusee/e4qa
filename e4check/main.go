package main

import (
	"fmt"
	"os"

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
	errs := check()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("%s\n", err.Error())
		}
		os.Exit(-1)
	}

}
