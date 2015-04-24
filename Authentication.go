// Authentication.go
package main

import (
	"bufio"
	"bytes"
	"deploy-engine/lib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func Authentication() bool {
	/*address := []string{"/cf-release/src/etcd/mod/dashboard/app/coreos-web/sass/compass/utilities/color/color",
		"/spec/assets/release/src/variant/haku/haku",
		"/cf-release/src/uaa/scim/src/test/java/org/cloudfoundry/identity/uaa/scim/bootstrap/bootstrap",
		"/cf-release/src/gorouter/Godeps/_workspace/src/code.google.com/p/gogoprotobuf/proto/testdata/testdata",
	}*/
	buf, err := ioutil.ReadFile(lib.MainPath + IDAddress)
	if err != nil {
		fmt.Println("File corrupted! Please buy a complete software.")
		return false
	}
	engineID = string(buf)
	//data := map[string]string{}
	data := map[string]string{}
	key := &bytes.Buffer{}
	if engineID == "" {
		for {
			if engineID == "" {
				fmt.Println("Please input product key:")
				bio := bufio.NewReader(os.Stdin)
				lin, _, _ := bio.ReadLine()
				engineID = strings.TrimRight(string(lin), string(10))
				//fmt.Println(engineID)
			} else {
				key.Write([]byte(engineID))
				resp, err := http.Post("http://183.129.190.82:50000/authentication", "text/html", key)
				if err != nil {
					fmt.Println("Please ensure the network is enable!")
					return false
				}
				buff := new(bytes.Buffer)
				buff.ReadFrom(resp.Body)

				_ = json.Unmarshal(buff.Bytes(), &data)
				if data["status"] == "0" {
					fmt.Println("The product key is not exist! Please buy a complete software.")
					engineID = ""
					continue
				} else if data["status"] == "2" || data["status"] == "1" {
					ioutil.WriteFile(lib.MainPath+IDAddress, []byte(data["engineID"]), 0664)
					ioutil.WriteFile(lib.MainPath+CountAddress, []byte(data["count"]), 0664)
					fmt.Println("Product activation success!")
					fmt.Println("Deployment Engine Server start success.")
					//fmt.Println(data["engineID"])
					//fmt.Println(data["count"])
					return true
				}
			}
		}
	}
	return true
}
