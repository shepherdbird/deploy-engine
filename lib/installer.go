// installer
package lib

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func NewInstaller(Jobname string, Index int, configpath string, templatepath string) *Installer {
	Conf := Config{}
	buf, _ := ioutil.ReadFile(configpath)
	_ = yaml.Unmarshal(buf, &Conf)
	prepare(configpath, templatepath)
	return &Installer{
		Jobname:    Jobname,
		Index:      Index,
		Conf:       &Conf,
		Temp:       templatepath,
		configpath: configpath,
	}
}
func prepare(configpath string, templatepath string) {
	Check(configpath, templatepath)
	Conf := Config1{}
	Temp := Template{}
	buf, _ := ioutil.ReadFile(configpath)
	_ = yaml.Unmarshal(buf, &Conf)
	buf, _ = ioutil.ReadFile(templatepath)
	_ = yaml.Unmarshal(buf, &Temp)
	ipmodifer := NewIPModifer(&Temp, &Conf, MainPath+"/manifests/cfyml")
	ipmodifer.Work()
	Work(configpath)
	adduser := NewAdduser(configpath)
	adduser.Work()
}
func (Ins *Installer) Run() {
	if Ins.Jobname != "" && Ins.Index >= 0 {
		auto := NewAutoInstall(Ins.Jobname, Ins.Index, Ins.Conf.Components[Ins.Jobname][Ins.Index], Ins.Conf.Properties["user"], Ins.Conf.Properties["password"], Ins.Temp)
		auto.Work()
	} else {
		mul := NewMultiThreadWork(Ins.configpath, Ins.Temp)
		mul.Work()
	}
}
