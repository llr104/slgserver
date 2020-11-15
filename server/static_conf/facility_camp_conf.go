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
var FCAMP facilityCampConf

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

type facilityCampConf struct {
	Title	string	`json:"title"`
	Han		camp	`json:"han"`
	Wei		camp	`json:"wei"`
	Shu		camp	`json:"shu"`
	Wu		camp	`json:"wu"`
	Qun		camp	`json:"qun"`
}

func (this *facilityCampConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_camp.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityCampConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}

func (this *facilityCampConf) MaxLevel(fType int8) int8 {
	if this.Han.Type == fType{
		return int8(len(this.Han.Levels))
	}else if this.Wei.Type == fType{
		return int8(len(this.Wei.Levels))
	}else if this.Shu.Type == fType{
		return int8(len(this.Shu.Levels))
	}else if this.Wu.Type == fType{
		return int8(len(this.Wu.Levels))
	}else if this.Qun.Type == fType{
		return int8(len(this.Qun.Levels))
	}else{
		return 0
	}
}

