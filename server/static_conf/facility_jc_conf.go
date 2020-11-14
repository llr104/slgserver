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

//城内设施校场
var FJC facilityJC

type jcLevel struct {
	Level	int8 `json:"level"`
	Cnt		int8 `json:"cnt"`
}

type facilityJC struct {
	Title	string				`json:"title"`
	Name	string				`json:"name"`
	Des		string				`json:"des"`
	Type	int8				`json:"type"`
	Levels	[]jcLevel			`json:"levels"`
}

func (this *facilityJC) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_jc.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityJC load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}