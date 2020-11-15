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

//城内设施集市
var FMarket facilityMarketConf

type arketLevel struct {
	Level	int8			`json:"level"`
	Rate	int8 			`json:"rate"`
	Need	levelNeedRes	`json:"need"`
}

type facilityMarketConf struct {
	Title	string				`json:"title"`
	Name	string				`json:"name"`
	Des		string				`json:"des"`
	Type	int8				`json:"type"`
	Levels	[]mbsLevel			`json:"levels"`
}

func (this *facilityMarketConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_market.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityMarketConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}

func (this *facilityMarketConf) MaxLevel(fType int8) int8 {
	if this.Type == fType{
		return int8(len(this.Levels))
	}else{
		return 0
	}
}