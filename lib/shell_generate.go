// shell_generate
package lib

import (
	//"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
)

func Work(configpath string) {
	config := Config{}
	buf, err := ioutil.ReadFile(configpath)
	if err != nil {
		SaveLog(err.Error(), "")
		//fmt.Println(err)
		panic(err)
	}
	err = yaml.Unmarshal(buf, &config)
	if !isDirExists(MainPath + "/scripts/alljobscripts") {
		SaveLog("no such file or directory.", "")
		//fmt.Println("no such file or directory.")
	} else {
		for comp, ips := range config.Components {
			for index, _ := range ips {
				filename := comp + "_" + strconv.Itoa(index)
				//fmt.Println(filename)
				f, err := os.OpenFile(MainPath+"/scripts/alljobscripts/"+filename+".sh", os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
				if err != nil {
					SaveLog(err.Error(), "")
					panic(err)
				}
				defer f.Close()
				f.WriteString("bash env.sh\n")
				f.WriteString("mv " + filename + "_cf.yml" + " cf.yml\n")
				f.WriteString("echo 'mv cf files'\n")
				f.WriteString("echo 'start exec autoinstall'\n")
				f.WriteString("bash autoinstall.sh\n")
				f.WriteString("cd nise_bosh/\n")
				f.WriteString("sudo bundle exec ./bin/nise-bosh --keep-monit-files -y -i " + strconv.Itoa(index) + " ../cf-release/ ../cf.yml " + comp + "\n")
				f.WriteString("sudo /var/vcap/bosh/bin/monit\n")
				//f.WriteString("sudo /var/vcap/bosh/bin/monit quit && sleep 2\n")
				//f.WriteString("sudo /var/vcap/bosh/bin/monit && sleep 2\n")
				//f.WriteString("sudo /var/vcap/bosh/bin/monit restart all\n")
			}
		}
		SaveLog("Generate all jobs install.sh complete.", "")
	}

}
func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
	panic("not reached")
}

//func main() {
//	work("/home/dawei/config.yml")
//}
