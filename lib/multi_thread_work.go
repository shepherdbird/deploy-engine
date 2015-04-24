package lib

import (
	//"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
	"sync"
)

func NewMultiThreadWork(config_file string, template string) *MultiThreadWork {
	c := Config{}
	buf, err := ioutil.ReadFile(config_file)
	if err != nil {
		SaveLog("no such file :"+err.Error(), "")
		panic("no such file :" + err.Error())
	}
	err = yaml.Unmarshal(buf, &c)
	return &MultiThreadWork{
		yml:          &c,
		threads:      []chan int{},
		install_list: []*AutoInstall{},
		ip_resource:  map[string]*sync.Mutex{},
		user:         c.Properties["user"],
		password:     c.Properties["password"],
		template:     template,
	}
}
func (mul *MultiThreadWork) addlist() {
	for job, ips := range mul.yml.Components {
		for index, host := range ips {
			mul.ip_resource[host] = &sync.Mutex{}
			mul.install_list = append(mul.install_list, NewAutoInstall(job, index, host, mul.user, mul.password, mul.template))
		}
	}
}
func (mul *MultiThreadWork) topsort_work() {
	for _, install_obj := range mul.install_list {
		pipe := make(chan int)
		mul.threads = append(mul.threads, pipe)
		go func(install_obj *AutoInstall, pipe chan int, ip_resource map[string]*sync.Mutex) {
			defer func() { // 必须要先声明defer，否则不能捕获到panic异常
				if err := recover(); err != nil {
					SaveLog("Job "+install_obj.job+" on "+install_obj.host+" ERROR.\n", "")
					SaveLog(err.(error).Error(), "")
					//fmt.Printf("Job: %s on %s  ERROR.\n", install_obj.job, install_obj.host)
					//fmt.Println(err)
					pipe <- 1
				}
			}()
			ip_resource[install_obj.host].Lock()
			SaveLog("Installing  "+install_obj.host+"  "+install_obj.job+"  "+strconv.Itoa(install_obj.index), "")
			//fmt.Printf("%s %s %d\n", install_obj.host, install_obj.job, install_obj.index)
			install_obj.Work()
			ip_resource[install_obj.host].Unlock()
			pipe <- 1
		}(install_obj, pipe, mul.ip_resource)
	}
	for _, pipe := range mul.threads {
		<-pipe
	}
}
func (mul *MultiThreadWork) Work() {
	mul.addlist()
	mul.topsort_work()
}
