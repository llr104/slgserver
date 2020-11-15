package static_conf

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

//城墙女墙配置
var FWALL facilityWallConf

type wallLevel struct {
	Level	int8			`json:"level"`
	Limit	int  			`json:"limit"`
	Need	levelNeedRes	`json:"need"`
}

type wall struct {
	Name	string          `json:"name"`
	Des		string          `json:"des"`
	Type	int8            `json:"type"`
	Levels	[]wallLevel		`json:"levels"`
}


type facilityWallConf struct {
	Title 	string		`json:"title"`
	CQ		wall		`json:"cq"`
	NQ		wall		`json:"nq"`
}

func (this *facilityWallConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_wall.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityWallConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}


