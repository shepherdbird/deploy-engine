// Verification.go
package main

import (
	//"bufio"
	"bytes"
	"deploy-engine/lib"
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"
	//"os"
	//"strings"
	"gopkg.in/yaml.v2"
)

func Verification() int {
	//lib.MainPath = "/home/dawei/cf_nise_installer"
	buf, err := ioutil.ReadFile(lib.MainPath + IDAddress)
	if err != nil {
		//lib.SaveLog("File corrupted! Please buy a complete software.", "")
		return 0
	}
	engineID = string(buf)
	buf, err = ioutil.ReadFile(lib.MainPath + CountAddress)
	if err != nil {
		//lib.SaveLog("File corrupted! Please buy a complete software.", "")
		return 0
	}
	Count = string(buf)
	data := map[string]string{
		"engineID": engineID,
		"Count":    Count,
	}
	//fmt.Println("%v", data)
	Jdata, _ := json.Marshal(data)
	key := &bytes.Buffer{}
	key.Write(Jdata)
	resp, err := http.Post("http://183.129.190.82:50000/check", "application/json", key)
	if err != nil {
		//fmt.Println("Please ensure the network is enable!")
		return 1
	}
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	_ = json.Unmarshal(buff.Bytes(), &data)
	//fmt.Println("%v", data)
	if data["engineID"] == "-1" && data["Count"] == "-1" {
		//fmt.Println("Conflict!! Maybe there is another machine with the same.")
		return 2
	}
	ioutil.WriteFile(lib.MainPath+IDAddress, []byte(data["engineID"]), 0664)
	ioutil.WriteFile(lib.MainPath+CountAddress, []byte(data["Count"]), 0664)
	//fmt.Println("verify success!")
	return 3
}

func Complete() int {
	//lib.MainPath = "/home/dawei/cf_nise_installer"
	buf, err := ioutil.ReadFile(lib.MainPath + IDAddress)
	if err != nil {
		//lib.SaveLog("File corrupted! Please buy a complete software.", "")
		return 0
	}
	engineID = string(buf)
	buf, err = ioutil.ReadFile(lib.MainPath + CountAddress)
	if err != nil {
		//lib.SaveLog("File corrupted! Please buy a complete software.", "")
		return 0
	}
	Count = string(buf)
	buf, err = ioutil.ReadFile(lib.MainPath + "/config/config.yml")
	conf := lib.Config{}
	if err == nil {
		yaml.Unmarshal(buf, &conf)
	}
	bd, _ := json.Marshal(conf)
	data := map[string]string{
		"engineID": engineID,
		"Count":    Count,
		"Config":   string(bd),
	}
	//fmt.Println("%v", data)
	Jdata, _ := json.Marshal(data)
	key := &bytes.Buffer{}
	key.Write(Jdata)
	resp, err := http.Post("http://183.129.190.82:50000/complete", "application/json", key)
	if err != nil {
		//fmt.Println("Please ensure the network is enable!")
		return 1
	}
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	_ = json.Unmarshal(buff.Bytes(), &data)
	//fmt.Println("%v", data)
	if data["engineID"] == "-1" && data["Count"] == "-1" {
		//fmt.Println("Conflict!! Maybe there is another machine with the same.")
		return 2
	}
	ioutil.WriteFile(lib.MainPath+IDAddress, []byte(data["engineID"]), 0664)
	ioutil.WriteFile(lib.MainPath+CountAddress, []byte(data["Count"]), 0664)
	//fmt.Println("verify success!")
	return 3
}
