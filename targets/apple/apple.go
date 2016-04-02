package apple

import (
	"log"

	"sevki.org/build/internal"
)

func init() {

	if err := internal.Register("ios_application", IOSApplication{}); err != nil {
		log.Fatal(err)
	}
}
