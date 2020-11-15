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

//城内设施募兵所
var FMBS facilityMBS

type mbsLevel struct {
	Level	int8			`json:"level"`
	Rate	int8 			`json:"rate"`
	Need	levelNeedRes	`json:"need"`
}

type facilityMBS struct {
	Title	string				`json:"title"`
	Name	string				`json:"name"`
	Des		string				`json:"des"`
	Type	int8				`json:"type"`
	Levels	[]mbsLevel			`json:"levels"`
}

func (this *facilityMBS) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_mbs.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facility_mbs load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}
