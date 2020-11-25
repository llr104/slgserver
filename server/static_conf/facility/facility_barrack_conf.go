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

//兵营预备役配置
var FBarrack facilityBarrackConf

type byLevel struct {
	Level int8    `json:"level"`
	Extra int     `json:"extra"`
	Need  NeedRes `json:"need"`
}

type by struct {
	Name	string      `json:"name"`
	Des		string   `json:"des"`
	Type	int8        `json:"type"`
	Levels	[]byLevel `json:"levels"`
}

type ybyLevel struct {
	Level int8    `json:"level"`
	Limit int     `json:"limit"`
	Need  NeedRes `json:"need"`
}

type yby struct {
	Name	string    `json:"name"`
	Des		string    `json:"des"`
	Type	int8      `json:"type"`
	Levels	[]ybyLevel`json:"levels"`
}


type facilityBarrackConf struct {
	Title string `json:"title"`
	BY    by     `json:"by"`
	YBY   yby    `json:"yby"`
	Types []int8 `json:"-"`
}

func (this *facilityBarrackConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility", "facility_barrack.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityBarrackConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	this.Types = make([]int8, 2)
	this.Types[0] = this.YBY.Type
	this.Types[1] = this.BY.Type

	fmt.Println(this)
}

func (this *facilityBarrackConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
}

func (this *facilityBarrackConf) MaxLevel(fType int8) int8 {
	if this.BY.Type == fType{
		return int8(len(this.BY.Levels))
	}else if this.YBY.Type == fType{
		return int8(len(this.YBY.Levels))
	}else{
		return 0
	}
}

func (this *facilityBarrackConf) Need(fType int8, level int) (*NeedRes, bool) {
	if level <= 0{
		return nil, false
	}

	if this.BY.Type == fType{
		if len(this.BY.Levels) >= level{
			return &this.BY.Levels[level-1].Need, true
		}else {
			return nil, false
		}

	}else if this.YBY.Type == fType{
		if len(this.YBY.Levels) >= level{
			return &this.YBY.Levels[level-1].Need, true
		}else {
			return nil, false
		}
	}else{
		return nil, false
	}
}


