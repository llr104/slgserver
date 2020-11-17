package static_conf

import (
	"encoding/json"
	"errors"
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

type LevelNeedRes struct {
	Decree 		int	`json:"decree"`
	Grain		int `json:"grain"`
	Wood		int `json:"wood"`
	Iron		int `json:"iron"`
	Stone		int `json:"stone"`
}

type produceResLevel struct {
	Level int8         `json:"level"`
	Yield int          `json:"yield"`
	Need  LevelNeedRes `json:"need"`
}

type produceRes struct {
	Name	string           `json:"name"`
	Des		string           `json:"des"`
	Type	int8             `json:"type"`
	Levels	[]produceResLevel`json:"levels"`
}

type facilityProduceResConf struct {
	Title string     	`json:"title"`
	FMC   produceRes 	`json:"fmc"`
	LTC   produceRes 	`json:"ltc"`
	MF    produceRes 	`json:"mf"`
	CSC   produceRes 	`json:"csc"`
	MJ    produceRes 	`json:"mj"`
	Types []int8		`json:"-"`
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

	this.Types = make([]int8, 5)
	this.Types[0] = this.FMC.Type
	this.Types[1] = this.CSC.Type
	this.Types[2] = this.LTC.Type
	this.Types[3] = this.MF.Type
	this.Types[4] = this.MJ.Type

	fmt.Println(this)
}

func (this *facilityProduceResConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
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

func (this *facilityProduceResConf) Need(fType int8, level int) (*LevelNeedRes, error)  {
	if this.CSC.Type == fType{
		if len(this.CSC.Levels) < level{
			return &this.CSC.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else if this.FMC.Type == fType{
		if len(this.FMC.Levels) < level{
			return &this.FMC.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else if this.LTC.Type == fType{
		if len(this.LTC.Levels) < level{
			return &this.LTC.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else if this.MF.Type == fType{
		if len(this.MF.Levels) < level{
			return &this.MF.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else if this.MJ.Type == fType{
		if len(this.MJ.Levels) < level{
			return &this.MJ.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else{
		return nil, errors.New("type not found")
	}
}
