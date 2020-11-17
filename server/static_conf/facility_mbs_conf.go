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

//城内设施募兵所
var FMBS facilityMBSConf

type mbsLevel struct {
	Level int8         `json:"level"`
	Rate  int8         `json:"rate"`
	Need  LevelNeedRes `json:"need"`
}

type facilityMBSConf struct {
	Title	string			`json:"title"`
	Name	string			`json:"name"`
	Des		string			`json:"des"`
	Type	int8			`json:"type"`
	Levels	[]mbsLevel		`json:"levels"`
	Types 	[]int8			`json:"-"`
}

func (this *facilityMBSConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_mbs.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facility_mbs_conf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	this.Types = make([]int8, 1)
	this.Types[0] = this.Type


	fmt.Println(this)
}

func (this *facilityMBSConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
}

func (this *facilityMBSConf) MaxLevel(fType int8) int8 {
	if this.Type == fType{
		return int8(len(this.Levels))
	}else{
		return 0
	}
}

func (this *facilityMBSConf) Need(fType int8, level int) (*LevelNeedRes, error)  {
	if this.Type == fType{
		if len(this.Levels) >= level{
			return &this.Levels[level-1].Need, nil
		}else{
			return nil, errors.New("level not found")
		}
	} else {
		return nil, errors.New("type not found")
	}
}