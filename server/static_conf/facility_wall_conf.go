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

//城墙女墙配置
var FWALL facilityWallConf

type wallLevel struct {
	Level int8         `json:"level"`
	Limit int          `json:"limit"`
	Need  LevelNeedRes `json:"need"`
}

type wall struct {
	Name	string          `json:"name"`
	Des		string          `json:"des"`
	Type	int8            `json:"type"`
	Levels	[]wallLevel		`json:"levels"`
}


type facilityWallConf struct {
	Title 	string		`json:"title"`
	CQ		wall		`json:"cq"`
	NQ		wall		`json:"nq"`
	Types 	[]int8		`json:"-"`
}

func (this *facilityWallConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility_wall.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityWallConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	this.Types = make([]int8, 2)
	this.Types[0] = this.NQ.Type
	this.Types[1] = this.CQ.Type

	fmt.Println(this)
}

func (this *facilityWallConf) IsContain(t int8) bool {
	for _, t1 := range this.Types {
		if t == t1 {
			return true
		}
	}
	return false
}

func (this *facilityWallConf) MaxLevel(fType int8) int8 {
	if this.CQ.Type == fType{
		return int8(len(this.CQ.Levels))
	}else if this.CQ.Type == fType{
		return int8(len(this.NQ.Levels))
	} else{
		return 0
	}
}

func (this *facilityWallConf) Need(fType int8, level int) (*LevelNeedRes, error)  {
	if this.CQ.Type == fType{
		if len(this.CQ.Levels) >= level{
			return &this.CQ.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}

	}else if this.NQ.Type == fType{
		if len(this.NQ.Levels) >= level{
			return &this.NQ.Levels[level].Need, nil
		}else {
			return nil, errors.New("level not found")
		}
	}else{
		return nil, errors.New("type not found")
	}
}
