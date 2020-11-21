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

//阵营配置
var FCAMP facilityCampConf

type campLevel struct {
	Level int8    `json:"level"`
	Rate  int     `json:"rate"`
	Need  NeedRes `json:"need"`
}

type camp struct {
	Name	string        `json:"name"`
	Des		string     `json:"des"`
	Type	int8          `json:"type"`
	Levels	[]campLevel `json:"levels"`
}

type facilityCampConf struct {
	Title string `json:"title"`
	Han   camp   `json:"han"`
	Wei   camp   `json:"wei"`
	Shu   camp   `json:"shu"`
	Wu    camp   `json:"wu"`
	Qun   camp   `json:"qun"`
	Types []int8 `json:"-"`
}

func (this *facilityCampConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility", "facility_camp.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityCampConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	this.Types = make([]int8, 5)
	this.Types[0] = this.Han.Type
	this.Types[1] = this.Wei.Type
	this.Types[2] = this.Shu.Type
	this.Types[3] = this.Wu.Type
	this.Types[4] = this.Qun.Type


	fmt.Println(this)
}

func (this *facilityCampConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
}

func (this *facilityCampConf) MaxLevel(fType int8) int8 {
	if this.Han.Type == fType{
		return int8(len(this.Han.Levels))
	}else if this.Wei.Type == fType{
		return int8(len(this.Wei.Levels))
	}else if this.Shu.Type == fType{
		return int8(len(this.Shu.Levels))
	}else if this.Wu.Type == fType{
		return int8(len(this.Wu.Levels))
	}else if this.Qun.Type == fType{
		return int8(len(this.Qun.Levels))
	}else{
		return 0
	}
}

func (this *facilityCampConf) Need(fType int8, level int) (*NeedRes, bool)  {
	if this.Han.Type == fType{
		if len(this.Han.Levels) > level{
			return &this.Han.Levels[level].Need, true
		}else {
			return nil, false
		}

	}else if this.Wei.Type == fType{
		if len(this.Wei.Levels) >= level{
			return &this.Wei.Levels[level-1].Need, true
		}else {
			return nil, false
		}
	}else if this.Shu.Type == fType{
		if len(this.Shu.Levels) >= level{
			return &this.Shu.Levels[level-1].Need, true
		}else {
			return nil, false
		}
	}else if this.Wu.Type == fType{
		if len(this.Wu.Levels) >= level{
			return &this.Wu.Levels[level-1].Need, true
		}else {
			return nil, false
		}
	}else if this.Qun.Type == fType{
		if len(this.Qun.Levels) >= level{
			return &this.Qun.Levels[level-1].Need, true
		}else {
			return nil, false
		}
	}else{
		return nil, false
	}
}
