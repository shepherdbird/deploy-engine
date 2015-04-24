package lib

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func NewCommand(path string, cmd string, job string, ip string) *Command {
	t := Config{}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
	}
	err = yaml.Unmarshal(buf, &t)
	return &Command{
		Yml:    &t,
		Cmd:    cmd,
		Job:    job,
		Ip:     ip,
		Result: []string{},
		Thread: []chan int{},
	}
}
func (c *Command) exec(key string, host string, user string, password string) {
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(password))
	config := &ssh.ClientConfig{
		User: user,
		Auth: auths,
	}
	//connect
	conn, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, 22), config)
	if e != nil {
		SaveLog("Fail to connect :"+e.Error(), "")
		panic("Fail to connect :" + e.Error())
	}
	c.Result = []string{}
	defer conn.Close()
	//exec
	ses, e := conn.NewSession()
	if e != nil {
		SaveLog("Fail to create session :"+e.Error(), "")
		panic("Fail to create session :" + e.Error())
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
	readl := NewReadl()
	ses.Stdout = readl
	go func() {
		w, _ := ses.StdinPipe()
		defer w.Close()
		for {
			select {
			case line, _ := <-readl.Out:
				SaveLog(line.String(), "")
				if strings.Contains(line.String(), "[sudo]") {
					fmt.Fprintln(w, password)
				} else {
					c.Result = append(c.Result, strings.Trim(line.String(), "\n"))
				}
				//fmt.Fprintln(w, "C0644", len(content), "testfile")
				//fmt.Fprintln(w, content)
				//fmt.Fprint(w, "\x00")
			}
		}

	}()
	if err := ses.Run(c.Cmd); err != nil {
		SaveLog("Fail to run :"+err.Error(), "")
		panic("Fail to run :" + err.Error())
	}
}
func (c *Command) Get_through() {
	for key, hosts := range c.Yml.Components {
		for _, host := range hosts {
			user := c.Yml.Properties["user"]
			password := c.Yml.Properties["password"]
			ch := make(chan int)
			c.Thread = append(c.Thread, ch)
			go func(key string, host string, user string, password string, ch chan int) {
				defer func() { // 必须要先声明defer，否则不能捕获到panic异常
					if err := recover(); err != nil {
						SaveLog("Job: "+key+" on "+host+" exec "+c.Cmd, "")
						//fmt.Printf("Job: %s on %s exec %s ERROR.\n", key, host, c.Cmd)
						SaveLog(err.(error).Error(), "")
						//fmt.Println(err)
						ch <- 1
					}
					//ch <- 1
					//fmt.Println("d")
				}()
				if c.Job != "" && key != c.Job {
					ch <- 0
					return
				}
				if c.Ip != "" && host != c.Ip {
					ch <- 0
					return
				}
				//c.thread = append(c.thread, ch)
				c.exec(key, host, user, password)
				ch <- 1

			}(key, host, user, password, ch)
		}
	}
	//fmt.Println(len(c.Thread))
	for _, v := range c.Thread {
		<-v
	}
}
func (c *Command) Show() {
	for i, v := range c.Result {
		fmt.Println(i)
		fmt.Println(v)
	}
	fmt.Println(len(c.Result))
}
func (c *Command) Status() string {
	for _, v := range c.Result {
		if strings.Contains(v, "not") || strings.Contains(v, "Excution") || strings.Contains(v, "Does") {
			return "DOWN"
		} else if strings.Contains(v, "initializing") {
			return "INITIALIZING"
		}
	}
	return "RUNNING"
}
