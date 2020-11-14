package static_conf

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"slgserver/config"
	"slgserver/log"
)

//城内生产资源设施配置
var FPRC facilityProduceResConf

type produceResLevel struct {
	Level      int8 `json:"level"`
	NeedDecree int8 `json:"need_decree"`
	NeedGrain  int  `json:"need_grain"`
	NeedWood   int  `json:"need_wood"`
	NeedIron   int  `json:"need_iron"`
	NeedStone  int  `json:"need_stone"`
	Yield      int  `json:"yield"`
}

type produceRes struct {
	Name	string           `json:"name"`
	Des		string           `json:"des"`
	Type	int8             `json:"type"`
	Levels	[]produceResLevel`json:"levels"`
}

type facilityProduceResConf struct {
	Title string     `json:"title"`
	FMC   produceRes `json:"fmc"`
	LTC   produceRes `json:"ltc"`
	MF    produceRes `json:"mf"`
	CSC   produceRes `json:"csc"`
	MJ    produceRes `json:"mj"`
}

func (this *facilityProduceResConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_produce_res.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facility_produce_res_conf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
}