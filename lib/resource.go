// resource
package lib

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	//"log"
	//"os"
	//"regexp"
	"strconv"
	"strings"
)

func Resourcework(configpath string) []Resource {
	var res []Resource
	ctrl := make(chan int, 1000)
	config := Config{}
	buf, err := ioutil.ReadFile(configpath)
	if err != nil {
		SaveLog(err.Error(), "")
		return nil
		//panic(err)
		//fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
	}
	err = yaml.Unmarshal(buf, &config)
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(config.Properties["root"]))
	conf := &ssh.ClientConfig{
		User: "root",
		Auth: auths,
	}
	for k, v := range config.Components {
		for _, j := range v {
			go func(k string, j string) {
				now_res := Resource{}
				now_res.Comp = k
				now_res.IP = j
				now_res.CPU = now_res.getCPU(j, conf)
				now_res.RAM = now_res.getRAM(j, conf)
				now_res.DISK = now_res.getDISK(j, conf) * 1024
				res = append(res, now_res)
				ctrl <- 0
			}(k, j)
		}
	}
	for _, v := range config.Components {
		for _, _ = range v {
			<-ctrl
		}
	}
	return res
}
func Resourcework_2(comp string, ip string, configpath string) Resource {
	now_res := Resource{}
	config := Config{}
	buf, err := ioutil.ReadFile(configpath)
	if err != nil {
		SaveLog(err.Error(), "")
		return now_res
		//panic(err)
		//fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
	}
	err = yaml.Unmarshal(buf, &config)
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(config.Properties["root"]))
	conf := &ssh.ClientConfig{
		User: "root",
		Auth: auths,
	}
	for c, ips := range config.Components {
		if c == comp && Exist(ips, ip) {
			now_res.Comp = comp
			now_res.IP = ip
			now_res.CPU = now_res.getCPU(ip, conf)
			now_res.RAM = now_res.getRAM(ip, conf)
			now_res.DISK = now_res.getDISK(ip, conf) * 1024
		}
	}
	SaveLog(comp+" and "+ip+" do not match!", "")
	return now_res
}
func (res *Resource) getCPU(ip string, conf *ssh.ClientConfig) int {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), conf)
	if err != nil {
		SaveLog("unable to connect: "+err.Error(), "")
		//log.Fatalf("unable to connect: %s", err)
		return 0
	}
	defer conn.Close()
	ses, e := conn.NewSession()
	if e != nil {
		SaveLog(e.Error(), "")
		return 0
	}
	var b bytes.Buffer
	ses.Stdout = &b
	ses.Run("lscpu | grep \"^CPU(s):\" | awk '{print $2}'")
	//fmt.Printf("CPU: %s", b.String())
	num, _ := strconv.Atoi(strings.Trim(b.String(), "\n"))
	return num
}
func (res *Resource) getRAM(ip string, conf *ssh.ClientConfig) int {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), conf)
	if err != nil {
		SaveLog("unable to connect: "+err.Error(), "")
		//log.Fatalf("unable to connect: %s", err)
		return 0
	}
	defer conn.Close()
	ses, e := conn.NewSession()
	if e != nil {
		SaveLog(e.Error(), "")
		return 0
	}

	var b bytes.Buffer
	ses.Stdout = &b
	ses.Run("grep MemTotal /proc/meminfo | awk '{print $2}'")
	//fmt.Printf("RAM: %s", b.String())
	num, _ := strconv.Atoi(strings.Trim(b.String(), "\n"))
	return num
}
func (res *Resource) getDISK(ip string, conf *ssh.ClientConfig) float64 {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), conf)
	if err != nil {
		SaveLog("unable to connect: "+err.Error(), "")
		//log.Fatalf("unable to connect: %s", err)
		return 0
	}
	defer conn.Close()
	ses, e := conn.NewSession()
	if e != nil {
		SaveLog(e.Error(), "")
		return 0
	}

	var b bytes.Buffer
	ses.Stdout = &b
	ses.Run("fdisk -l | grep \"GB\" | awk '{print $3}'")
	//fmt.Printf("DISK: %s", b.String())
	num, _ := strconv.ParseFloat(strings.Trim(b.String(), "\n"), 10)
	return num
}

/*func main() {
	//work("/home/dawei/config.yml")
	//total_resource := []Resource{}
	tt := work("/home/dawei/config.yml")
	fmt.Printf("%v", tt)
}*/
