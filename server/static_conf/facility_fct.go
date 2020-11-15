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
var FFCT facilityFCT


type fctLevel struct {
	Level	int8			`json:"level"`
	Limit	int  			`json:"limit"`
	Need	levelNeedRes	`json:"need"`
}


type facilityFCT struct {
	Title 	string		`json:"title"`
	Name	string		`json:"name"`
	Des		string		`json:"des"`
	Type	int8		`json:"type"`
	Levels	[]fctLevel	`json:"levels"`
}

func (this *facilityFCT) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_fct.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityFCT load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}


