package apple

import (
	"log"

	"sevki.org/build/ast"
)

func init() {

	if err := ast.Register("ios_application", IOSApplication{}); err != nil {
		log.Fatal(err)
	}
}
