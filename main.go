// deploy-engine project main.go
package main

import (
	"bytes"
	"deploy-engine/lib"
	//"deploy-engine/test"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	//"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

//var Lread, Lwrite *os.File
var flag = make(chan int)

//var temporary = make([]byte, 200)
var JobToProcess = map[string]string{
	"nats":                "nats",
	"haproxy":             "haproxy",
	"nfs_server":          "rpc",
	"database":            "postgres",
	"cloud_controller_ng": "cloud_controller",
	"etcd":                "etcd",
	"uaa":                 "uaa",
	"login":               "login",
	"hm9000":              "hm9000",
	"loggregator":         "loggregator",
	"collector":           "collector",
	"syslog_aggregator":   "syslog_aggregator",
	"gorouter":            "gorouter",
	"dea":                 "dea",
}

func config(w http.ResponseWriter, req *http.Request) {
	lib.LogFileName = "error.log"
	if req.Method == "GET" {
		buf, err := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
		conf := lib.Config{}
		if err == nil {
			yaml.Unmarshal(buf, &conf)
		}
		bd, _ := json.Marshal(conf)
		w.Write([]byte(bd))
		w.WriteHeader(http.StatusOK)
	} else if req.Method == "POST" || req.Method == "PUT" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		//fmt.Printf("%v", buf)
		config := lib.Config{}
		err := json.Unmarshal(buf.Bytes(), &config)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			//log.Fatalf("error: %v", err)
			lib.SaveLog(err.Error(), "")
			w.Write([]byte(err.Error()))
		} else {
			d, err := yaml.Marshal(&config)
			if err != nil {
				//log.Fatalf("error: %v", err)
				lib.SaveLog(err.Error(), "")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else {
				ioutil.WriteFile(lib.MainPath+"/config/config.yml", d, 0644)
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("Save success!"))
			}
		}
	}
}
func Log1(ws *websocket.Conn) {
	//fmt.Println("hello world!")
	/*go func() {
		for {
			Lread.Read(temporary)
			_ = websocket.Message.Send(ws, strings.TrimRight(string(temporary), string(0)))
		}

	}()*/
	go func() {
		for {
			select {
			case log, _ := <-lib.Log:
				//fmt.Println("%v", log)
				for {
					err := websocket.Message.Send(ws, log)
					if err == nil {
						break
					}
				}

			}
		}

	}()
	<-flag
}
func deploy(w http.ResponseWriter, req *http.Request) {
	lib.LogFileName = "error.log"
	switch Verification() {
	case 0:
		lib.SaveLog("File corrupted! Please buy a complete software.", "")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("File corrupted! Please buy a complete software."))
	case 1:
		lib.SaveLog("Please ensure the network is enable!", "")
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("Please ensure the network is enable!"))
	case 2:
		lib.SaveLog("Sorry.You are allow to deploy PaaS 50 times.", "")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Sorry.You are allow to deploy PaaS 50 times."))
	case 3:
		lib.LogFileName = fmt.Sprintf(time.Now().UTC().String()) + "_deploy.log"
		w.WriteHeader(http.StatusAccepted)
		lib.NewInstaller("", -1, lib.MainPath+"/config/config.yml", lib.MainPath+"/manifests/template.yml").Run()
		Complete()
		ioutil.WriteFile(lib.MainPath+"/config/deploy", []byte("1"), 0x664)
	}
}
func status(w http.ResponseWriter, req *http.Request) {
	lib.LogFileName = "status.log"
	ctrl := make(chan int, 1000)
	buf, err := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
	if err != nil {
		w.Write([]byte("Please complete the configuration file."))
		w.WriteHeader(http.StatusForbidden)
		lib.SaveLog("Please complete the configuration file.", "")
		return
	}
	if lib.Fideploy == "0" {
		w.Write([]byte("Please deploy your cluster firstly."))
		w.WriteHeader(http.StatusForbidden)
		lib.SaveLog("Please deploy your cluster firstly.", "")
		return
	}
	conf := lib.Config{}
	_ = yaml.Unmarshal(buf, &conf)
	res := []JobStatus{}

	for comp, ips := range conf.Components {
		for _, ip := range ips {
			go func(comp string, ip string) {
				var lin JobStatus
				cmd := lib.NewCommand(lib.MainPath+"/config/config.yml", "sudo /var/vcap/bosh/bin/monit summary | grep "+JobToProcess[comp]+" | awk '{print $3}'", comp, ip)
				cmd.Get_through()
				lin.work(comp, ip, &conf)
				lin.Status = cmd.Status()
				res = append(res, lin)
				ctrl <- 0
			}(comp, ip)
		}
	}
	for _, ips := range conf.Components {
		for _, _ = range ips {
			<-ctrl
		}
	}
	for _, re := range res {
		fmt.Printf("%s %s %s %s %s %s\n", re.Comp, re.Ip, re.Cpu, re.Mem, re.Disk, re.Status)
		lib.SaveLog(fmt.Sprintf("%s %s %s %s %s %s", re.Comp, re.Ip, re.Cpu, re.Mem, re.Disk, re.Status), "")
	}

	bd, _ := json.Marshal(res)
	//w.Header().Add("Content-Type", "aplication/json")
	w.Write([]byte(bd))
	w.WriteHeader(http.StatusOK)
}
func restart(w http.ResponseWriter, req *http.Request) {
	lib.LogFileName = "error.log"
	ctrl := make(chan int, 1000)
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	bd := RestartInfo{}
	err := json.Unmarshal(buf.Bytes(), &bd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		lib.SaveLog(err.Error(), "")
		//log.Fatalf("error: %v", err)
		w.Write([]byte(err.Error()))
		return
	}
	if lib.Fideploy == "0" {
		w.Write([]byte("Please deploy your cluster firstly."))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if bd.Name == "all" && bd.Ip == "0.0.0.0" {
		buf, err := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
		if err != nil {
			w.Write([]byte("Please complete the configuration file."))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		conf := lib.Config{}
		_ = yaml.Unmarshal(buf, &conf)
		usedip := []string{}
		for comp, ips := range conf.Components {
			for _, ip := range ips {
				if !lib.Exist(usedip, ip) {
					go func(comp string, ip string) {
						cmd := lib.NewCommand(lib.MainPath+"/config/config.yml", "sudo /var/vcap/bosh/bin/monit restart all", comp, ip)
						cmd.Get_through()
						ctrl <- 0
					}(comp, ip)
					//cmd.Show()
				}
				usedip = append(usedip, ip)
			}
		}
		for _, ips := range conf.Components {
			for _, _ = range ips {
				<-ctrl
			}
		}
	} else {
		buf, err := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
		if err != nil {
			w.Write([]byte("Please complete the configuration file."))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		conf := lib.Config{}
		_ = yaml.Unmarshal(buf, &conf)
		for comp, ips := range conf.Components {
			if bd.Name == comp {
				if lib.Exist(ips, bd.Ip) {
					cmd := lib.NewCommand(lib.MainPath+"/config/config.yml", "sudo /var/vcap/bosh/bin/monit restart all", bd.Name, bd.Ip)
					cmd.Get_through()
					w.WriteHeader(http.StatusOK)
					return
				} else {
					lib.SaveLog(bd.Name+" and "+bd.Ip+" do not match!", "")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(bd.Name + " and " + bd.Ip + " do not match!\n"))
				}
			}
		}
	}
}
func stop(w http.ResponseWriter, req *http.Request) {
	lib.LogFileName = "error.log"
	ctrl := make(chan int, 1000)
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	bd := RestartInfo{}
	err := json.Unmarshal(buf.Bytes(), &bd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		lib.SaveLog(err.Error(), "")
		//log.Fatalf("error: %v", err)
		w.Write([]byte(err.Error()))
		return
	}
	if lib.Fideploy == "0" {
		w.Write([]byte("Please deploy your cluster firstly."))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if bd.Name == "all" && bd.Ip == "0.0.0.0" {
		buf, _ := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
		if err != nil {
			w.Write([]byte("Please complete the configuration file."))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		conf := lib.Config{}
		_ = yaml.Unmarshal(buf, &conf)
		usedip := []string{}
		for comp, ips := range conf.Components {
			for _, ip := range ips {
				if !lib.Exist(usedip, ip) {
					go func(comp string, ip string) {
						cmd := lib.NewCommand(lib.MainPath+"/config/config.yml", "sudo /var/vcap/bosh/bin/monit stop all", comp, ip)
						cmd.Get_through()
						ctrl <- 0
					}(comp, ip)
					//cmd.Show()
					usedip = append(usedip, ip)
				}
			}
		}
		for _, ips := range conf.Components {
			for _, _ = range ips {
				<-ctrl
			}
		}
	} else {
		buf, err := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
		if err != nil {
			w.Write([]byte("Please complete the configuration file."))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		conf := lib.Config{}
		_ = yaml.Unmarshal(buf, &conf)
		for comp, ips := range conf.Components {
			if bd.Name == comp {
				if lib.Exist(ips, bd.Ip) {
					cmd := lib.NewCommand(lib.MainPath+"/config/config.yml", "sudo /var/vcap/bosh/bin/monit stop all", bd.Name, bd.Ip)
					cmd.Get_through()
					w.WriteHeader(http.StatusOK)
					return
				} else {
					lib.SaveLog(bd.Name+" and "+bd.Ip+" do not match!", "")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(bd.Name + " and " + bd.Ip + " do not match!\n"))
				}
			}
		}
	}
}
func update(w http.ResponseWriter, req *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	bd := map[string][]string{}
	err := json.Unmarshal(buf.Bytes(), &bd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatalf("error: %v", err)
		w.Write([]byte(err.Error()))
		return
	}
	buff, _ := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
	conf := lib.Config{}
	_ = yaml.Unmarshal(buff, &conf)
	for comp, ips := range bd {
		for _, ip := range ips {
			for Lcomp, Lips := range conf.Components {
				if Lcomp == comp {
					if !lib.Exist(Lips, ip) {
						Lips = append(Lips, ip)
						d, _ := yaml.Marshal(&conf)
						ioutil.WriteFile(lib.MainPath+"/config/config.yml", d, 0664)
						lib.NewInstaller(comp, len(Lips)-1, lib.MainPath+"/config/config.yml", lib.MainPath+"/manifests/template.yml").Run()
					}
				}
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}
func download(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		logs := []string{}
		dirs, _ := ioutil.ReadDir(lib.MainPath + "/log/")
		for _, dir := range dirs {
			logs = append(logs, dir.Name())

		}
		js, _ := json.Marshal(logs)
		//fmt.Println(logs)
		w.Write(js)
		w.WriteHeader(http.StatusOK)
	} else if req.Method == "POST" || req.Method == "PUT" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		log := buf.String()
		fmt.Println(log)
		fmt.Println(lib.MainPath + "/log/" + strings.TrimRight(log, "\n"))
		_, err := os.Stat(lib.MainPath + "/log/" + strings.TrimRight(log, "\n"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("File do not exist."))
			return
		}
		http.ServeFile(w, req, lib.MainPath+"/log/"+strings.TrimRight(log, "\n"))
		//http.Redirect(w, req, lib.MainPath+"/log/"+strings.TrimRight(log, "\n"), http.StatusFound)
	}
}
func hardware(w http.ResponseWriter, req *http.Request) {
	lib.LogFileName = "hardware.log"
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	bd := map[string]string{}
	json.Unmarshal(buf.Bytes(), &bd)
	if bd["Comp"] == "all" && bd["Ip"] == "0.0.0.0" {
		res := lib.Resourcework(lib.MainPath + "/config/config.yml")
		lib.SaveLog(fmt.Sprintf("%v", res), "")
		resp, _ := json.Marshal(res)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	} else {
		cf := lib.Config{}
		buff, err := ioutil.ReadFile(lib.MainPath + "/config/config.yml")
		if err != nil {
			lib.SaveLog(err.Error(), "")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please complete the configuration file.\n"))
			return
			//panic(err)
			//fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
		}
		yaml.Unmarshal(buff, &cf)
		for c, k := range cf.Components {
			if c == bd["Comp"] {
				if lib.Exist(k, bd["Ip"]) {
					res := lib.Resourcework_2(bd["Comp"], bd["Ip"], lib.MainPath+"/config/config.yml")
					lib.SaveLog(fmt.Sprintf("%v", res), "")
					resp, _ := json.Marshal(res)
					w.WriteHeader(http.StatusOK)
					w.Write(resp)
					return
				} else {
					lib.SaveLog(bd["Comp"]+" and "+bd["Ip"]+" do not match!", "")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(bd["Comp"] + " and " + bd["Ip"] + " do not match!\n"))
				}

			}
		}
	}
}
func main() {
	//Lread, Lwrite, _ = os.Pipe()
	//os.Stdout = Lwrite
	if !Authentication() {
		return
	}
	http.HandleFunc("/deployment-engine/config", config)
	http.HandleFunc("/deployment-engine/update", update)
	http.HandleFunc("/deployment-engine/deploy", deploy)
	http.HandleFunc("/deployment-engine/status", status)
	http.HandleFunc("/deployment-engine/restart", restart)
	http.HandleFunc("/deployment-engine/stop", stop)
	http.HandleFunc("/deployment-engine/download", download)
	http.HandleFunc("/deployment-engine/hardware", hardware)
	http.Handle("/deployment-engine/log", websocket.Handler(Log1))
	err := http.ListenAndServe(lib.ControllerIp+":50000", nil)
	//err := http.ListenAndServe("127.0.0.1:50000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
	//test.AdduserTest()
	//test.CflogerTest()
	//test.ShellGenerateTest()
	//test.ResourceTest()
	//test.ReleaseTest() //cannot read the size of files
	//test.CheckTest()
	//test.IpmodifierTest() //cannot change domain!
	//test.AutoInstallTest()
	//test.MultiWorkTest()
	//test.CommandTest()
	//fmt.Println(lib.MainPath)
	//lib.NewInstaller("", -1, lib.MainPath+"/config/config.yml", lib.MainPath+"/manifests/template.yml").Run()
}
func init() {
	run := exec.Command("/bin/bash", "-c", "pwd")
	stdout, _ := run.StdoutPipe()
	run.Start()
	out, _ := ioutil.ReadAll(stdout)
	lib.MainPath = strings.TrimRight(string(out), string(10))
	//lib.MainPath = "/home/dawei/cf_nise_installer"

	run = exec.Command("/bin/bash", "-c", "ifconfig eth0 | grep \"inet addr:\" | awk '{print $2}' | cut -d \":\" -f 2")
	stdout, _ = run.StdoutPipe()
	run.Start()
	out, _ = ioutil.ReadAll(stdout)
	lib.ControllerIp = strings.TrimRight(string(out), string(10))

	run = exec.Command("/bin/bash", "-c", "whoami")
	stdout, _ = run.StdoutPipe()
	run.Start()
	out, _ = ioutil.ReadAll(stdout)
	lib.ControllerUser = strings.TrimRight(string(out), string(10))
	buf, err := ioutil.ReadFile(lib.MainPath + "/config/deploy")
	if err != nil {
		fmt.Println("File corrupted! Please buy a complete software.")
	}
	lib.Fideploy = strings.TrimRight(string(buf), string(10))
}
