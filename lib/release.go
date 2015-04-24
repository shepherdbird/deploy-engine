// release
package lib

import (
	//"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

func NewRelease(job string, ssh *Scp, templatepath string) *Release {
	deploy_manifest := Template{}
	buf, err := ioutil.ReadFile(templatepath)
	if err != nil {
		SaveLog(err.Error(), "")
		panic(err)
	}
	err = yaml.Unmarshal(buf, &deploy_manifest)
	var templates []string
	templates = nil
	for _, m := range deploy_manifest.Jobs {
		if m.Name == job {
			templates = m.Template
		}
	}
	release_dir := MainPath + "/cf-release/"
	release_file := initReleaseFile(release_dir)
	packages := getPackages(templates, deploy_manifest, release_file)
	return &Release{
		Job:             job,
		Deploy_manifest: &deploy_manifest,
		Templates:       templates,
		Release_dir:     release_dir,
		Packages:        packages,
		Release_file:    release_file,
		Ssh:             ssh,
	}
}
func initReleaseFile(release_dir string) *ReleaseFile {
	configdir := release_dir + "config/"
	newest_release := "0"
	//final yml
	finalconfigpath := configdir + "final.yml"
	finalnameL := make(map[string]string)
	finalname := ""
	buf, err := ioutil.ReadFile(finalconfigpath)
	if err == nil {
		_ = yaml.Unmarshal(buf, &finalnameL)
		finalname = finalnameL["final_name"]
	}
	//index yml
	finalindexpath := release_dir + "releases/index.yml"
	finalindex := make(map[string]string)
	finalindexL := FinalIndexL{}
	buf, err = ioutil.ReadFile(finalindexpath)
	if err == nil {
		_ = yaml.Unmarshal(buf, &finalindexL)
		for _, finalindex = range finalindexL.Builds {
			ver, _ := strconv.Atoi(finalindex["version"])
			nw, _ := strconv.Atoi(newest_release)
			if ver > nw {
				newest_release = strconv.Itoa(ver)
			}
		}
	} else {
		finalindex = nil
	}
	// dev yml
	devconfigpath := configdir + "dev.yml"
	devnameL := make(map[string]string)
	devname := ""
	buf, err = ioutil.ReadFile(devconfigpath)
	if err == nil {
		_ = yaml.Unmarshal(buf, &devnameL)
		devname = devnameL["dev_name"]
	}
	//dev_releases index yml
	devindexpath := release_dir + "dev_releases/cf/index.yml"
	devindex := make(map[string]string)
	devindexL := FinalIndexL{}
	buf, err = ioutil.ReadFile(devindexpath)
	if err == nil {
		_ = yaml.Unmarshal(buf, &devindexL)
		for _, devindex = range devindexL.Builds {
			if strings.Contains(newest_release, "dev") {
				nw := strings.Split(newest_release, "+")
				dv := strings.Split(devindex["version"], "+")
				nwL, _ := strconv.Atoi(nw[0])
				dvL, _ := strconv.Atoi(dv[0])
				if dvL > nwL {
					newest_release = devindex["version"]
				} else if dvL == nwL {
					nwM := strings.Split(nw[1], ".")
					dvM := strings.Split(dv[1], ".")
					nwMM, _ := strconv.Atoi(nwM[1])
					dvMM, _ := strconv.Atoi(dvM[1])
					if dvMM > nwMM {
						newest_release = devindex["version"]
					}
				}
			} else {
				dv := strings.Split(devindex["version"], "+")
				nwL, _ := strconv.Atoi(newest_release)
				dvL, _ := strconv.Atoi(dv[0])
				if dvL >= nwL {
					newest_release = devindex["version"]
				}
			}
		}
	} else {
		devindex = nil
	}

	if finalindex == nil && devindex == nil {
		SaveLog("No release index found!\nTry `bosh create release` in your release repository.", "")
	}
	release_file := ""
	if strings.Contains(newest_release, "dev") {
		release_file = release_dir + "dev_releases/cf/" + devname + "-" + newest_release + ".yml"
	} else {
		release_file = release_dir + "releases/" + finalname + "-" + newest_release + ".yml"
	}
	var r *ReleaseFile
	buf, err = ioutil.ReadFile(release_file)
	if err == nil {
		_ = yaml.Unmarshal(buf, &r)
	} else {
		panic(err)
	}
	return r
}
func getPackages(templates []string, deploy_manifest Template, release_file *ReleaseFile) []string {
	pak := []string{""}
	release_dir := MainPath + "/cf-release/"
	for _, template := range templates {
		var version string
		for _, k := range release_file.Jobs {
			if k.Name == template {
				version = k.Version
				break
			}
		}
		job_tgz_path := release_dir + ".final_builds/jobs/" + template + "/" + version + ".tgz"
		job_MF := make(map[string][]string)
		cmd := exec.Command("/bin/bash", "-c", "tar -Ozxf "+job_tgz_path+" ./job.MF")
		stdout, _ := cmd.StdoutPipe()
		_ = cmd.Start()
		bytes, _ := ioutil.ReadAll(stdout)
		_ = yaml.Unmarshal(bytes, &job_MF)
		for _, j := range job_MF["packages"] {
			if !Exist(pak, j) {
				pak = append(pak, j)
			}
		}
	}
	return pak
}
func Exist(a []string, b string) bool {
	for _, k := range a {
		if k == b {
			return true
		}
	}
	return false
}

func (R *Release) Build() {
	myssh := NewMyscp(R.Ssh)
	myssh.exec("mkdir ~/cf-release")

	R.Ssh.Upload(R.Release_dir+"config", "~/cf-release/")
	R.Ssh.Upload(R.Release_dir+"packages", "~/cf-release/")
	R.Ssh.Upload(R.Release_dir+"releases", "~/cf-release/")
	myssh.exec("mkdir ~/cf-release/.final_builds")
	R.Ssh.Upload(R.Release_dir+".final_builds/jobs", "~/cf-release/.final_builds/")
	myssh.exec("mkdir ~/cf-release/.final_builds/packages")
	p := R.Packages
	final_p := []string{}
	start := 0
	last := len(p) - 1
	for start <= last {
		if !Exist(final_p, p[start]) {
			final_p = append(final_p, p[start])
		}
		for _, k := range R.Release_file.Packages {
			if k.Name == p[start] {
				for _, j := range k.Dependencies {
					p = append(p, j)
					last++
				}
				break
			}
		}
		start++
	}
	for _, l := range final_p {
		if l != "" {
			src := R.Release_dir + ".final_builds/packages/" + l
			dst := "~/cf-release/.final_builds/packages/"
			R.Ssh.Upload(src, dst)
			//fmt.Println(src)
		}

	}
}
