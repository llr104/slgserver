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

//兵营预备役配置
var FBarrack facilityBarrack

type byLevel struct {
	Level	int8			`json:"level"`
	Extra	int  			`json:"extra"`
	Need	levelNeedRes	`json:"need"`
}

type by struct {
	Name	string          `json:"name"`
	Des		string          `json:"des"`
	Type	int8            `json:"type"`
	Levels	[]byLevel		`json:"levels"`
}

type ybyLevel struct {
	Level	int8			`json:"level"`
	Limit	int  			`json:"limit"`
	Need	levelNeedRes	`json:"need"`
}

type yby struct {
	Name	string          `json:"name"`
	Des		string          `json:"des"`
	Type	int8            `json:"type"`
	Levels	[]ybyLevel		`json:"levels"`
}


type facilityBarrack struct {
	Title 	string		`json:"title"`
	BY		by			`json:"by"`
	YBY		yby			`json:"yby"`
}

func (this *facilityBarrack) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_barrack.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityBarrack load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}


