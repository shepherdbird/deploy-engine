package lib

import (
	//"bytes"
	"fmt"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"

	//"golang.org/x/crypto/ssh/agent"
	//"golang.org/x/crypto/ssh/terminal"
	"os/exec"
	"strings"
)

func (s *Scp) Upload1(local string, dst string) {

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
		//log.Fatalf("request for pseudo terminal failed: %s", err)
	}
	readl := NewReadl()
	ses.Stdout = readl
	go func() {
		w, _ := ses.StdinPipe()
		defer w.Close()
		//content := "123456"
		for {
			select {
			case line, _ := <-readl.Out:
				SaveLog(strings.TrimRight(line.String(), string(10)), "")
				if strings.Contains(line.String(), "password") {
					fmt.Fprintln(w, s.Localpwd)
				} else if strings.Contains(line.String(), "Are you sure you want to continue connecting") {
					fmt.Fprintln(w, "yes")
				}
				//fmt.Fprintln(w, "C0644", len(content), "testfile")
				//fmt.Fprintln(w, content)
				//fmt.Fprint(w, "\x00")
			}
		}

	}()
	if err := ses.Run("scp -r " + s.Localuser + "@" + s.Localip + ":" + local + " " + dst); err != nil {
		SaveLog("Fail to run :"+err.Error(), "")
		panic("Fail to run :" + err.Error())
	}
}
func (s *Scp) Upload(local string, dst string) {
	var ch = make(chan int)
	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("scp -r %s %s@%s:%s ", local, s.User, s.Host, dst))
	//fmt.Println("1")
	f, err := pty.Start(c)
	if err != nil {
		//fmt.Println(err.Error())
		SaveLog(err.Error(), "")
	}
	go func() {
		for {
			select {
			case <-ch:
				return
			default:
			}
			var b = make([]byte, 1000)
			f.Read(b)
			//fmt.Printf("%s", string(b))
			SaveLog(strings.TrimRight(string(b), string(0)), "")
			if strings.Contains(string(b), "continue") {
				f.WriteString("yes\n")
			}
			if strings.Contains(string(b), "password") {
				f.WriteString(s.Password + "\n")
				//time.Sleep(time.Second)
			}
		}
	}()
	err = c.Wait()
	ch <- 1
	if err != nil {
		fmt.Println(err.Error())
	}
}

/*func main() {
	scp := &Scp{
		localuser: "zjw",
		localpwd:  "123456",
		localip:   "10.10.105.158",
	}
	scp.upload("/home/zjw/adduser.rb", "~/")
	fi, err := os.Stat("../../goproject")
	if err != nil {
		fmt.Println("YY")
	}
	if fi.IsDir() {
		fmt.Println("NN")
	}

}*/
