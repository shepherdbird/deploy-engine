// CommandTest.go
package test

import (
	"deploy-engine/lib"
	"fmt"
)

func CommandTest() {
	cmd := lib.NewCommand(lib.MainPath+"/config/config.yml", "sudo /var/vcap/bosh/bin/monit summary | grep "+JobToProcess[comp]+" | awk '{print $3}'", "", "")
	cmd.Get_through()
	cmd.Show()
	fmt.Println("Hello World!")
}
