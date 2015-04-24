// IpmodifierTest.go
package test

import (
	"deploy-engine/lib"
	//"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func IpmodifierTest() {
	Conf := lib.Config1{}
	Temp := lib.Template{}
	buf, _ := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
	_ = yaml.Unmarshal(buf, &Conf)
	buf, _ = ioutil.ReadFile(lib.MainPath + "/manifests/template.yml")
	_ = yaml.Unmarshal(buf, &Temp)
	ipmodifer := lib.NewIPModifer(&Temp, &Conf, lib.MainPath+"/manifests/cfyml")
	ipmodifer.Work()
	//fmt.Println("Hello World!")
}
