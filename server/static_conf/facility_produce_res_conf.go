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

//城内生产资源设施配置
var FPRC facilityProduceResConf

type levelNeedRes struct {
	Decree 		int	`json:"decree"`
	Grain		int `json:"grain"`
	Wood		int `json:"wood"`
	Iron		int `json:"iron"`
	Stone		int `json:"stone"`
}

type produceResLevel struct {
	Level	int8			`json:"level"`
	Yield	int  			`json:"yield"`
	Need	levelNeedRes	`json:"need"`
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
	fmt.Println(this)
}

func (this *facilityProduceResConf) MaxLevel(fType int8) int8 {
	if this.CSC.Type == fType{
		return int8(len(this.CSC.Levels))
	}else if this.FMC.Type == fType{
		return int8(len(this.FMC.Levels))
	}else if this.LTC.Type == fType{
		return int8(len(this.LTC.Levels))
	}else if this.MF.Type == fType{
		return int8(len(this.MF.Levels))
	}else if this.MJ.Type == fType{
		return int8(len(this.MJ.Levels))
	}else{
		return 0
	}
}