package lib

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
)

func template_has(template *Template, jobname string) bool {
	for _, v := range template.Jobs {
		if v.Name == jobname {
			return true
		}
	}
	return false
}
func Check(configpath string, templatepath string) bool {
	//read file
	config := Config{}
	template := Template{}
	buf, err := ioutil.ReadFile(configpath)
	if err != nil {
		SaveLog(err.Error(), "")
		return false
	}
	err = yaml.Unmarshal(buf, &config)
	buf, err = ioutil.ReadFile(templatepath)
	if err != nil {
		SaveLog(err.Error(), "")
		return false
	}
	err = yaml.Unmarshal(buf, &template)
	//config vs template
	for k, _ := range config.Components {
		if !template_has(&template, k) {
			SaveLog("Your manifest/template don't have this job: "+k, "")
			return false
		}
	}
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(config.Properties["root"]))
	conf := &ssh.ClientConfig{
		User: "root",
		Auth: auths,
	}
	for _, v := range config.Components {
		for _, j := range v {
			conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", j, 22), conf)
			if err != nil {
				SaveLog("unable to connect: "+err.Error(), "")
				return false
			}
			defer conn.Close()
		}
	}
	i, _ := strconv.Atoi(config.Properties["checkdisk"])
	if i == 0 {
		SaveLog("Warn: config checkdisk option 0. It means don't check.", "")
		return true
	} else {
		now_res := Resource{}
		for k, v := range config.Components {
			for _, j := range v {
				now_res.Comp = k
				now_res.IP = j
				now_res.DISK = now_res.getDISK(j, conf)
				lin, _ := strconv.Atoi(config.Properties["disksize"])
				if now_res.DISK < float64(lin) {
					SaveLog(now_res.IP+"'s VM disk size is not Enough. use 'df -h' to check the VM disk.", "")
					return false
				} else {
					SaveLog(strconv.Itoa(int(now_res.DISK))+"GB >= "+strconv.Itoa(lin)+"GB         OK", "")
				}
			}
		}
	}
	return true
}
