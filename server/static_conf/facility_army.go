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

//四营配置
var FARMY facilityArmy

type armyLevel struct {
	Level	int8			`json:"level"`
	Rate	int  			`json:"rate"`
	Need	levelNeedRes	`json:"need"`
}

type army struct {
	Name	string          `json:"name"`
	Des		string          `json:"des"`
	Type	int8            `json:"type"`
	Levels	[]armyLevel		`json:"levels"`
}

type facilityArmy struct {
	Title string	`json:"title"`
	JFY   army		`json:"jfy"`
	TBY   army		`json:"tby"`
	JJY   army		`json:"jjy"`
	SWY   army		`json:"swy"`
}

func (this *facilityArmy) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_army.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityArmy load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}


