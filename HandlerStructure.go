// status.go
package main

import (
	"bytes"
	"deploy-engine/lib"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
)

var engineID, Count string
var IDAddress = "/cf-release/src/etcd/mod/dashboard/app/coreos-web/sass/compass/utilities/color/color"
var CountAddress = "/cf-release/src/uaa/scim/src/test/java/org/cloudfoundry/identity/uaa/scim/bootstrap/bootstrap"

type JobStatus struct {
	Comp   string
	Ip     string
	Cpu    string
	Mem    string
	Disk   string
	Status string
}
type RestartInfo struct {
	Name string
	Ip   string
}
type UpdateInfo struct {
	Components []string
}

func (JS *JobStatus) getCPU(ip string, conf *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), conf)
	if err != nil {
		lib.SaveLog("unable to connect: "+err.Error(), "")
		return ""
	}
	defer conn.Close()
	ses, e := conn.NewSession()
	if e != nil {
		lib.SaveLog(e.Error(), "")
		return ""
	}
	var b bytes.Buffer
	ses.Stdout = &b
	ses.Run("cat /proc/stat | grep 'cpu ' | awk '{print ($2+$3+$4)/($2+$3+$4+$5)*100}'")
	//fmt.Println(b.String())
	return strings.Trim(b.String(), "\n") + "%"
}
func (JS *JobStatus) getRAM(ip string, conf *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), conf)
	if err != nil {
		lib.SaveLog("unable to connect: "+err.Error(), "")
		//log.Fatalf("unable to connect: %s", err)
		return ""
	}
	defer conn.Close()
	ses, e := conn.NewSession()
	if e != nil {
		lib.SaveLog(e.Error(), "")
		return ""
	}

	var b bytes.Buffer
	ses.Stdout = &b
	ses.Run("free -m | grep 'Mem:' | awk '{print $3/$2*100}'")
	//fmt.Printf("RAM: %s", b.String())
	num := strings.Trim(b.String(), "\n")
	return num[:5] + "%"
}
func (JS *JobStatus) getDISK(ip string, conf *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), conf)
	if err != nil {
		lib.SaveLog("unable to connect: "+err.Error(), "")
		//log.Fatalf("unable to connect: %s", err)
		return ""
	}
	defer conn.Close()
	ses, e := conn.NewSession()
	if e != nil {
		lib.SaveLog(e.Error(), "")
		return ""
	}

	var b bytes.Buffer
	ses.Stdout = &b
	ses.Run("df -h | grep ' /' | awk 'NR==1 {print $5}'")
	//fmt.Printf("DISK: %s", b.String())
	//num, _ := strconv.ParseFloat(strings.Trim(b.String(), "\n"), 10)
	return strings.Trim(b.String(), "\n")
}
func (JS *JobStatus) work(comp string, ip string, config *lib.Config) {
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(config.Properties["root"]))
	conf := &ssh.ClientConfig{
		User: "root",
		Auth: auths,
	}
	JS.Comp = comp
	JS.Ip = ip
	JS.Cpu = JS.getCPU(ip, conf)
	JS.Mem = JS.getRAM(ip, conf)
	JS.Disk = JS.getDISK(ip, conf)
}
