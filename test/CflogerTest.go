// cflogertest.go
package test

import (
	"deploy-engine/lib"
	//"fmt"
)

func CflogerTest() {
	cont := "hello world"
	lib.SaveLog(cont, "")
	//fmt.Println("Hello World!")
}
