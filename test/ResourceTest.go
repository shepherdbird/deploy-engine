// ResourceTest.go
package test

import (
	"deploy-engine/lib"
	"fmt"
)

func ResourceTest() {
	res := lib.Resourcework(lib.MainPath + "/config.yml")
	fmt.Printf("%v", res)
}
