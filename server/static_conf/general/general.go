package general

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"slgserver/config"
	"slgserver/log"
)

var General general

type g struct {
	Name     string `json:"name"`
	CfgId    int    `json:"cfgId"`
	Force    int    `json:"force"`
	Strategy int    `json:"strategy"`
	Defense  int    `json:"defense"`
	Speed    int    `json:"speed"`
	Destroy  int    `json:"destroy"`
	Cost     int    `json:"cost"`
}

type general struct {
	Title	string	`json:"title"`
	List	[]g
}

func (this *general) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "general", "general.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("general load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	fmt.Println(this)
}