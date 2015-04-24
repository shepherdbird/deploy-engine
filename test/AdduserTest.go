// addusertest.go
package test

import (
	"deploy-engine/lib"
	//"fmt"
)

func AdduserTest() {
	t := lib.NewAdduser("./config.yml")
	t.Work()
	//fmt.Println("Hello World!")
}
