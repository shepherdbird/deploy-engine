// MultiWorkTest.go
package test

import (
	"deploy-engine/lib"
	"fmt"
)

func MultiWorkTest() {
	lib.NewMultiThreadWork("/home/dawei/mygo/src/deploy-engine/config.yml", lib.MainPath+"/manifests/template.yml").Work()
	fmt.Println("Hello World!")
}
