// CheckTest.go
package test

import (
	"deploy-engine/lib"
	"fmt"
)

func CheckTest() {
	fmt.Println(lib.Check(lib.MainPath+"/config/config.yml", lib.MainPath+"/manifests/template.yml"))
}
