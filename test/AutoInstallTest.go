// AutoInstallTest.go
package test

import (
	"deploy-engine/lib"
	"fmt"
)

func AutoInstallTest() {
	lib.NewAutoInstall("nats", 0, "10.10.101.181", "vcap", "password", lib.MainPath+"/manifests/template.yml").Work()
	fmt.Println("Hello World!")
}
