package lib

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

func NewAutoInstall(job string, index int, host string, user string, password string, temp string) *AutoInstall {
	return &AutoInstall{
		job:      job,
		index:    index,
		host:     host,
		user:     user,
		password: password,
		temp:     temp,
		filename: job + "_" + strconv.Itoa(index),
	}
}
func (auto *AutoInstall) send_file(s *Myssh, srcpath string, despath string) {
	_, err := os.Stat(srcpath)
	if err != nil {
		SaveLog("ERROR: Don't have such File: "+srcpath+"\n"+err.Error(), "")
		panic("ERROR: Don't have such File: " + srcpath + "\n" + err.Error())
	} else {
		s.Scpcli.Upload(srcpath, despath)
	}
}
func (auto *AutoInstall) send_blobs(s *Myssh, srcpath string, despath string) {
	fi, err := os.Stat(srcpath)
	if err != nil {
		SaveLog("ERROR: Don't have such File: "+srcpath+"\n"+err.Error(), "")
		panic("ERROR: Don't have such File: " + srcpath + "\n" + err.Error())
	} else {
		_, filename := path.Split(srcpath)
		filesize := fi.Size()
		if strings.Contains(s.exec("ls"), filename) {
			size, _ := strconv.Atoi(strings.TrimRight(s.exec("du -b "+filename+" | awk '{print $1}'"), "\n\r"))
			if int64(size) == filesize {
				return
			}
		}
		s.Scpcli.Upload(srcpath, despath)
	}
}
func (auto *AutoInstall) send_all(s *Myssh) {
	var file []string
	basepath := MainPath
	file = append(file, path.Join(basepath, "config", "sources.list"))
	file = append(file, path.Join(basepath, "manifests", "cfyml", auto.filename+"_cf.yml"))
	file = append(file, path.Join(basepath, "scripts", "alljobscripts", auto.filename+".sh"))
	file = append(file, path.Join(basepath, "scripts", "autoinstall.sh"))
	file = append(file, path.Join(basepath, "scripts", "env.sh"))
	file = append(file, path.Join(basepath, "scripts", "monit_start.sh"))

	var blobs []string
	srcpath := MainPath + "/blobs"
	dir, err := ioutil.ReadDir(srcpath)
	if err != nil {
		SaveLog("no dir:"+err.Error(), "")
		panic("no dir:" + err.Error())
	}
	for _, fi := range dir {
		blobs = append(blobs, path.Join(basepath, "blobs", fi.Name()))
	}
	for _, f := range file {
		auto.send_file(s, f, ".")
	}
	for _, f := range blobs {
		auto.send_blobs(s, f, ".")
	}
	NewRelease(auto.job, s.Scpcli, auto.temp).Build()

}
func (auto *AutoInstall) exec(s *Myssh, instructor string) {
	ses, e := s.Client.NewSession()
	if e != nil {
		panic("Fail to create session :" + e.Error())
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := ses.RequestPty("xterm", 80, 40, modes); err != nil {
		panic("Fail to create session :" + err.Error())
	}
	readl := NewReadl()
	ses.Stdout = readl
	go func() {
		in, _ := ses.StdinPipe()
		defer in.Close()
		for {
			select {
			case line, _ := <-readl.Out:
				SaveLog(strings.TrimRight(line.String(), string(10)), "")
				if strings.Contains(line.String(), "[sudo]") {
					fmt.Fprintln(in, auto.password)
				} else if strings.Contains(line.String(), "password") {
					fmt.Fprintln(in, auto.password)
				} else if strings.Contains(line.String(), "[]") {
					fmt.Fprintln(in, "")
				} else if strings.Contains(line.String(), "[Y/n]") {
					fmt.Fprintln(in, "y")
				} else if strings.Contains(line.String(), "[y/N]") {
					fmt.Fprintln(in, "y")
				} else {
					fmt.Printf("%v", line.String())
				}
			}
		}

	}()
	if err := ses.Run(instructor); err != nil {
		SaveLog(err.Error(), "")
		panic("Fail to run :" + err.Error())
	}
}
func (auto *AutoInstall) remote_connect(host string, user string, password string, localip string, localuser string, localpw string) {
	myssh := NewMyssh(host, user, password, localip, localuser, localpw)
	auto.send_all(myssh)
	auto.exec(myssh, "bash "+auto.filename+".sh")

}
func (auto *AutoInstall) Work() {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			SaveLog(auto.job+" install ERROR.", "")
			SaveLog(err.(error).Error(), "")
		}
	}()
	conf := Config{}
	buf, _ := ioutil.ReadFile(MainPath + "/config/config.yml")
	_ = yaml.Unmarshal(buf, &conf)
	auto.remote_connect(auto.host, auto.user, auto.password, ControllerIp, ControllerUser, conf.Properties["root"])
}
