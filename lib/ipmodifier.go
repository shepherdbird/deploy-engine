package lib

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func NewIPModifer(template *Template, config *Config1, path string) *IPModifer {
	return &IPModifer{
		TemplateObj: template,
		ConfigObj:   config,
		Path:        path,
	}
}
func (m *IPModifer) modifyIp() {
	ipMap := m.ConfigObj.Components
	if ipMap["database"] == nil {
		ipMap["database"][0] = "127.0.0.1"
	}
	if ipMap["ccdb"] == nil {
		ipMap["ccdb"] = ipMap["database"]
	}
	if ipMap["uaadb"] == nil {
		ipMap["uaadb"] = ipMap["database"]
	}
	ipMap["ccdb_ng"] = ipMap["database"]
	ipMap["databases"] = ipMap["database"]
	ipMap["router"] = ipMap["gorouter"]
	ipMap["loggregator_endpoint"] = ipMap["loggregator"]
	proMap := m.TemplateObj.Properties
	for component, ipSet := range ipMap {
		switch proMap[component].(type) {
		case map[interface{}]interface{}:
			pr := proMap[component].(map[interface{}]interface{})
			if pr["address"] != nil {
				pr["address"] = ipSet[0]
			}
			if pr["network"] != nil {
				pr["network"] = ipSet[0].(string) + "/24"
			}
			if pr["machines"] != nil {
				pr["machines"] = ipSet
			}
			if pr["host"] != nil {
				pr["host"] = ipSet[0]
			}
			if pr["servers"] != nil {
				p := pr["servers"].(map[interface{}]interface{})
				if p["z1"] != nil {
					p["z1"] = ipSet
				}
				if p["default"] != nil {
					p["default"] = ipSet
				}
			}
		}
	}
	delete(ipMap, "ccdb_ng")
	delete(ipMap, "databases")
	delete(ipMap, "router")
	delete(ipMap, "loggregator_endpoint")
}
func (m *IPModifer) modifyDomain(obj interface{}, oldDomain string, newDomain string) interface{} {
	switch obj.(type) {
	case Job:
		obj_temp := obj.(Job)
		obj_temp.Template = m.modifyDomain(obj_temp.Template, oldDomain, newDomain).([]string)
	case *Template:
		obj_temp := obj.(*Template)
		obj_temp.Jobs = m.modifyDomain(obj_temp.Jobs, oldDomain, newDomain).([]Job)
		obj_temp.Properties = m.modifyDomain(obj_temp.Properties, oldDomain, newDomain).(map[string]interface{})
	case map[interface{}]interface{}:
		obj_temp := obj.(map[interface{}]interface{})
		for k, v := range obj_temp {
			obj_temp[k] = m.modifyDomain(v, oldDomain, newDomain)
		}
	case []interface{}:
		obj_temp := obj.([]interface{})
		for k, v := range obj_temp {
			obj_temp[k] = m.modifyDomain(v, oldDomain, newDomain)
		}
	case string:
		obj = strings.Replace(obj.(string), oldDomain, newDomain, -1)
	default:
		fmt.Printf("%v\n\n\n\n", obj)
	}
	return obj
}
func (m *IPModifer) modifyDB(component string) {
	if component == "ccdb" || component == "uaadb" {
		m.TemplateObj.Properties["db"] = component
	} else {
		m.TemplateObj.Properties["db"] = "databases"
	}
}
func (m *IPModifer) modifyNats_Stream_forward(component string, index int) {
	if component == "nats" {
		natsinfo := m.TemplateObj.Properties["nats"].(map[interface{}]interface{})
		natsinfo["address"] = natsinfo["machines"].([]interface{})[index]
	}
}
func (m *IPModifer) output(dirpath string) bool {
	fi, err := os.Stat(dirpath)
	if err != nil {
		os.MkdirAll(dirpath, 0777)
	} else if !fi.IsDir() {
		os.MkdirAll(dirpath, 0777)
	}
	//fmt.Println("sss")
	//fmt.Println("%v", m.ConfigObj.Components)
	for component, ipSet := range m.ConfigObj.Components {
		for i, _ := range ipSet {
			m.modifyDB(component)
			m.modifyNats_Stream_forward(component, i)
			out, _ := yaml.Marshal(m.TemplateObj)
			err := ioutil.WriteFile(path.Join(m.Path, component+"_"+strconv.Itoa(i)+"_cf.yml"), out, 0777)
			if err != nil {
				SaveLog(err.Error(), "")
				//fmt.Println("%v", component)
				panic(err)
			}
		}
	}
	return true
}
func (m *IPModifer) Work() {
	newDomain := m.ConfigObj.Properties["domain"]
	oldDomain := m.TemplateObj.Properties["domain"].(string)
	rep, _ := regexp.Compile(`<.+>`)
	if rep.FindString(newDomain) != "" {
		comp := rep.FindString(newDomain)
		comp1 := strings.TrimLeft(comp, "<")
		comp1 = strings.TrimRight(comp, ">")
		ip := m.ConfigObj.Components[comp1][0].(string)
		newDomain = strings.Replace(newDomain, comp, ip, -1)
	}
	//fmt.Printf("%v\n", newDomain)
	tem, _ := yaml.Marshal(m.TemplateObj)
	tem = []byte(strings.Replace(string(tem), oldDomain, newDomain, -1))
	yaml.Unmarshal(tem, m.TemplateObj)
	//m.TemplateObj = m.modifyDomain(m.TemplateObj, oldDomain, newDomain).(*Template)
	m.modifyIp()
	m.output(m.Path)
}
