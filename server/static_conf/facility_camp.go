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

//阵营配置
var FCAMP facilityCamp

type campLevel struct {
	Level	int8			`json:"level"`
	Rate	int  			`json:"rate"`
	Need	levelNeedRes	`json:"need"`
}

type camp struct {
	Name	string          `json:"name"`
	Des		string          `json:"des"`
	Type	int8            `json:"type"`
	Levels	[]campLevel		`json:"levels"`
}

type facilityCamp struct {
	Title	string	`json:"title"`
	Han		camp	`json:"han"`
	Wei		camp	`json:"wei"`
	Shu		camp	`json:"shu"`
	Wu		camp	`json:"wu"`
	Qun		camp	`json:"qun"`
}

func (this *facilityCamp) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_camp.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityCamp load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}


