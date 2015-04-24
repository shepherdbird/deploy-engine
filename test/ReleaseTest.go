// ReleaseTest.go
package test

import (
	"deploy-engine/lib"
	//"fmt"
)

func ReleaseTest() {
	s := lib.Scp{"root", "123456", "10.10.102.249", "dawei", "123456", "10.10.105.111"}
	re := lib.NewRelease("nats", &s, lib.MainPath+"/manifests/template.yml")
	re.Build()
	//fmt.Println("Hello World!")
}
