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

//社稷坛配置
var FSJT facilitySJTConf


type sjtLevel struct {
	Level	int8			`json:"level"`
	Limit	int  			`json:"limit"`
	Need	levelNeedRes	`json:"need"`
}


type facilitySJTConf struct {
	Title 	string		`json:"title"`
	Name	string		`json:"name"`
	Des		string		`json:"des"`
	Type	int8		`json:"type"`
	Levels	[]sjtLevel	`json:"levels"`
}

func (this *facilitySJTConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_sjt.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilitySJTConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}


