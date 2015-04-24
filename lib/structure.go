// structure
package lib

import (
	//"fmt"
	"bytes"
	"golang.org/x/crypto/ssh"
	"sync"
)

var Log = make(chan string, 10000)
var MainPath string
var ControllerIp, ControllerUser string
var LogFileName = "error.log"
var Fideploy string

type Config struct {
	Components map[string][]string
	Properties map[string]string
}
type Adduser struct {
	Yml          *Config
	Rootpassword string
	User         string
	Userpassword string
}
type Readl struct {
	Out chan bytes.Buffer
}

type Job struct {
	Name     string
	Template []string
}
type Template struct {
	Deployment string
	Jobs       []Job
	Properties map[string]interface{}
}

type Command struct {
	Yml    *Config
	Thread []chan int
	Cmd    string
	Result []string
	Job    string
	Ip     string
}
type Config1 struct {
	Components map[string][]interface{}
	Properties map[string]string
}
type IPModifer struct {
	TemplateObj *Template
	ConfigObj   *Config1
	Path        string
}

type Scp struct {
	User      string
	Password  string
	Host      string
	Localuser string
	Localpwd  string
	Localip   string
}

type Resource struct {
	Comp string
	IP   string
	CPU  int
	RAM  int
	DISK float64
}

type Pack struct {
	Name         string
	Version      string
	Sha1         string
	Fingerprint  string
	Dependencies []string
}
type Jb struct {
	Name        string
	Version     string
	Fingerprint string
	Sha1        string
}
type ReleaseFile struct {
	Packages            []Pack
	Jobs                []Jb
	Commit_hash         string
	Uncommitted_changes string
	Name                string
	Version             string
}
type Release struct {
	Job             string
	Deploy_manifest *Template
	Templates       []string
	Release_dir     string
	Packages        []string
	Release_file    *ReleaseFile
	Ssh             *Scp
}
type FinalIndexL struct {
	Builds map[string](map[string]string)
}
type Myssh struct {
	User     string
	Password string
	Ip       string
	Client   *ssh.Client
	Scpcli   *Scp
}
type Installer struct {
	Jobname    string
	Index      int
	Conf       *Config
	Temp       string
	configpath string
}
type MultiThreadWork struct {
	yml          *Config
	threads      []chan int
	install_list []*AutoInstall
	ip_resource  map[string]*sync.Mutex
	user         string
	password     string
	template     string
}

type AutoInstall struct {
	job      string
	index    int
	host     string
	user     string
	password string
	temp     string
	filename string
}
