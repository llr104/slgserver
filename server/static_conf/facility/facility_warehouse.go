package facility

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
var FWareHouse facilityWarehouseConf

type wareHouseLevel struct {
	Level int8    `json:"level"`
	Limit int     `json:"limit"`
	Need  NeedRes `json:"need"`
}

type facilityWarehouseConf struct {
	Title	string      	`json:"title"`
	Name	string       	`json:"name"`
	Des		string    		`json:"des"`
	Type	int8         	`json:"type"`
	Levels	[]wareHouseLevel`json:"levels"`
	Types 	[]int8      	`json:"-"`
}

func (this *facilityWarehouseConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility", "facility_warehouse.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityWarehouseConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	this.Types = make([]int8, 1)
	this.Types[0] = this.Type

	fmt.Println(this)
}

func (this *facilityWarehouseConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
}

func (this *facilityWarehouseConf) MaxLevel(fType int8) int8 {
	if this.Type == fType{
		return int8(len(this.Levels))
	}else{
		return 0
	}
}

func (this *facilityWarehouseConf) Need(fType int8, level int) (*NeedRes, bool)  {
	if this.Type == fType{
		if len(this.Levels) >= level{
			return &this.Levels[level-1].Need, true
		}else{
			return nil, false
		}
	} else {
		return nil, false
	}
}

func (this *facilityWarehouseConf) Limit(level int) int{
	if level >= len(this.Levels){
		return 10000
	}else{
		return this.Levels[level-1].Limit
	}
}