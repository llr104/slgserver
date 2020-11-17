package facility

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

//四营配置
var FARMY facilityArmyConf

type armyLevel struct {
	Level int8         `json:"level"`
	Rate  int          `json:"rate"`
	Need  LevelNeedRes `json:"need"`
}

type army struct {
	Name	string        `json:"name"`
	Des		string     `json:"des"`
	Type	int8          `json:"type"`
	Levels	[]armyLevel `json:"levels"`
}

type facilityArmyConf struct {
	Title string `json:"title"`
	JFY   army   `json:"jfy"`
	TBY   army   `json:"tby"`
	JJY   army   `json:"jjy"`
	SWY   army   `json:"swy"`
	Types []int8 `json:"-"`
}

func (this *facilityArmyConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility", "facility_army.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityArmyConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	this.Types = make([]int8, 4)
	this.Types[0] = this.JFY.Type
	this.Types[1] = this.JJY.Type
	this.Types[2] = this.SWY.Type
	this.Types[3] = this.TBY.Type

	fmt.Println(this)
}

func (this *facilityArmyConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
}

func (this *facilityArmyConf) MaxLevel(fType int8) int8 {
	if this.JFY.Type == fType{
		return int8(len(this.JFY.Levels))
	}else if this.JJY.Type == fType{
		return int8(len(this.JJY.Levels))
	}else if this.SWY.Type == fType{
		return int8(len(this.SWY.Levels))
	}else if this.TBY.Type == fType{
		return int8(len(this.TBY.Levels))
	}else{
		return 0
	}
}

func (this *facilityArmyConf) Need(fType int8, level int) (*LevelNeedRes, error)  {
	if this.JFY.Type == fType{
		if len(this.JFY.Levels) >= level{
			return &this.JFY.Levels[level-1].Need, nil
		}else {
			return nil, errors.New("level not found")
		}

	}else if this.JJY.Type == fType{
		if len(this.JJY.Levels) >= level{
			return &this.JJY.Levels[level-1].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else if this.SWY.Type == fType{
		if len(this.SWY.Levels) >= level{
			return &this.SWY.Levels[level-1].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else if this.TBY.Type == fType{
		if len(this.TBY.Levels) >= level{
			return &this.TBY.Levels[level-1].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else{
		return nil, errors.New("type not found")
	}
}