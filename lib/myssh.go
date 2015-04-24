package lib

import (
	//"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	//"gopkg.in/yaml.v2"
	//"io/ioutil"
	"log"
	//"os"
	//"strconv"
	"strings"
)

func NewMyssh(ip string, user string, password string, localip string, localuser string, localpwd string) *Myssh {
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(password))
	config := &ssh.ClientConfig{
		User: user,
		Auth: auths,
	}
	conn, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, 22), config)
	if e != nil {
		//log.Fatalf("unable to connect: %s", e)
		SaveLog(e.Error(), "")
		panic(e.Error())
	}
	return &Myssh{
		Client:   conn,
		User:     user,
		Password: password,
		Ip:       ip,
		Scpcli: &Scp{
			User:      user,
			Password:  password,
			Host:      ip,
			Localuser: localuser,
			Localpwd:  localpwd,
			Localip:   localip,
		},
	}
}
func NewMyscp(s *Scp) *Myssh {
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.Password(s.Password))
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: auths,
	}
	conn, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, 22), config)
	if e != nil {
		SaveLog(e.Error(), "")
		//log.Fatalf("unable to connect: %s", e)
		panic(e.Error())
	}
	return &Myssh{
		Client:   conn,
		User:     s.User,
		Password: s.Password,
		Ip:       s.Host,
		Scpcli: &Scp{
			User:      s.User,
			Password:  s.Password,
			Host:      s.Host,
			Localuser: s.Localuser,
			Localpwd:  s.Localpwd,
			Localip:   s.Localip,
		},
	}
}
func (m *Myssh) exec(cmd string) string {
	var res string

	ses, e := m.Client.NewSession()
	if e != nil {
		return ""
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
				SaveLog(strings.TrimRight(line.String(), string(10)), "")
				res = res + line.String()
				if strings.Contains(line.String(), "[sudo]") {
					fmt.Fprintln(w, m.Password)
				} else {
					//c.result[key+host] = append(c.result[key+host], strings.Trim(line.String(), "\n"))
				}
				//fmt.Fprintln(w, "C0644", len(content), "testfile")
				//fmt.Fprintln(w, content)
				//fmt.Fprint(w, "\x00")
			}
		}

	}()
	//fmt.Println(host)
	if err := ses.Run(cmd); err != nil {
		SaveLog(err.Error(), "")
		//fmt.Println("%v", err.Error())
		//panic("Fail to run :" + err.Error())
	}
	return res
}

/*func main() {
	mys := NewMyssh("10.10.102.249", "root", "123456", "10.10.105.158", "zjw", "123456")
	z := mys.exec("du -b adduser.rb | awk '{print $1}'")
	fmt.Println("%v", []byte(z))
	fmt.Printf("%v", string(13) == "\n")

	z = strings.TrimRight(z, "\n\r")

	fmt.Printf("%v", []byte(z))
	t, err := strconv.Atoi(z)
	if err != nil {
		fmt.Println("tttt")
	}
	fmt.Printf("%d", t)
}*/
