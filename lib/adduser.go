package lib

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func (r *Readl) Write(p []byte) (n int, err error) {
	t := bytes.Buffer{}
	t.Write(p)
	r.Out <- t
	return len(p), nil
}
func NewReadl() *Readl {
	t := make(chan bytes.Buffer, 1000)
	return &Readl{
		Out: t,
	}
}
func NewAdduser(configpath string) *Adduser {
	t := Config{}
	buf, err := ioutil.ReadFile(configpath)
	if err != nil {
		SaveLog(err.Error(), "")
		return nil
	}
	err = yaml.Unmarshal(buf, &t)
	rootpassword := t.Properties["root"]
	user := t.Properties["user"]
	userpassword := t.Properties["password"]
	return &Adduser{
		Yml:          &t,
		Rootpassword: rootpassword,
		User:         user,
		Userpassword: userpassword,
	}
}
func (t *Adduser) useradd(ip string, config *ssh.ClientConfig) {

	conn, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), config)
	if e != nil {
		SaveLog(e.Error(), "")
		return
	}
	defer conn.Close()

	ses, e := conn.NewSession()
	if e != nil {
		SaveLog(e.Error(), "")
		return
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := ses.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}

	b := NewReadl()
	//in, e := ses.StdinPipe()
	//var out = make(chan bytes.Buffer, 10)
	//var b bytes.Buffer
	//var c []byte
	ses.Stdout = b
	in, e := ses.StdinPipe()
	//out, e := ses.StdoutPipe()
	go func() {
		for true {
			select {
			case line, _ := <-b.Out:
				SaveLog(strings.TrimRight(line.String(), string(10)), "")
				if strings.Contains(line.String(), "new UNIX password") {
					fmt.Fprintln(in, t.Userpassword)
				} else if strings.Contains(line.String(), "Is the information correct? [Y/n]") {
					fmt.Fprintln(in, "y")
				} else if strings.Contains(line.String(), "[]") {
					fmt.Fprintln(in, "")
				} else if strings.Contains(line.String(), "[Y/n]") {
					fmt.Fprintln(in, "y")
				} else if strings.Contains(line.String(), "[y/N]") {
					fmt.Fprintln(in, "y")
				} else if strings.Contains(line.String(), "already exists") {
				} else if strings.Contains(line.String(), "already a") {
				}
			}

			//fmt.Fscanln(out, b)
			//fmt.Printf("%v\n", b.String())
			//fmt.Println(b.String())
			//fmt.Fprintln(in, "123456")
			//b.String()
			//time.Sleep(100000000)
		}
	}()
	ses.Run("adduser " + t.User + ";adduser " + t.User + " sudo")
}
func (t *Adduser) Work() {
	isadduser, _ := strconv.Atoi(t.Yml.Properties["adduser"])
	if isadduser == 0 {
		SaveLog("Warn: config adduser 0. It means don't add.", "")
		return
	}
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(t.Rootpassword))
	config := &ssh.ClientConfig{
		User: "root",
		Auth: auths,
	}
	for _, ips := range t.Yml.Components {
		for _, ip := range ips {
			t.useradd(ip, config)
		}
	}
}
