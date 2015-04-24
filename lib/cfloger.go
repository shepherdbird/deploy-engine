// cfloger
package lib

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func SaveLog(content string, filepath string) {

	fi, err := os.Stat(MainPath + "/log")
	if err != nil {
		os.MkdirAll(MainPath+"/log", 0666)
	} else if !fi.IsDir() {
		os.MkdirAll(MainPath+"/log", 0666)
	}
	filepath = MainPath + "/log/" + LogFileName
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		Log <- fmt.Sprintf(err.Error() + "\n")
		//return
		//panic(err)
	}
	defer f.Close()
	Log <- fmt.Sprintf(strings.TrimRight(content, string(10)) + "\n")
	f.WriteString("[" + time.Now().String() + "] " + strings.TrimRight(content, string(10)) + "\n")
}
