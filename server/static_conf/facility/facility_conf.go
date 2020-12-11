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


const (
	Main 			= 0		//主城
	JiaoChang		= 13	//校场
	TongShuaiTing	= 14	//统帅厅
	MBS				= 16	//募兵所
)

var FConf facilityConf

type conf struct {
	Name	string
	Type	int8
}

type facilityConf struct {
	Title		string 				`json:"title"`
	List 		[]conf  			`json:"list"`
	facilitys 	map[int8]*facility
}

func (this *facilityConf) Load()  {
	this.facilitys = make(map[int8]*facility, 0)
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility", "facility.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityConf load file error",
			zap.Error(err),
			zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	fdir := path.Join(jsonDir, "facility")
	files, err := ioutil.ReadDir(fdir)
	if err != nil{
		return
	}
	for _, file := range files{
		if file.IsDir(){
			continue
		}else{
			if file.Name() == "facility.json" || file.Name() == "facility_addition.json"{
				continue
			}
			fileName := path.Join(fdir, file.Name())
			f := NewFacility(fileName)
			this.facilitys[f.Type] = f
		}
	}
	fmt.Println(this)
}

func (this *facilityConf) MaxLevel(fType int8) int8 {
	f, ok := this.facilitys[fType]
	if ok {
		return int8(len(f.Levels))
	}else{
		return 0
	}
}

func (this *facilityConf) Need(fType int8, level int8) (*NeedRes, bool) {
	if level <= 0{
		return nil, false
	}

	f, ok := this.facilitys[fType]
	if ok {
		if int8(len(f.Levels)) >= level {
			return &f.Levels[level-1].Need, true
		}else{
			return nil, false
		}
	}else{
		return nil, false
	}
}

func (this *facilityConf) GetValues(fType int8, level int8) []int {
	if level <= 0{
		return []int{}
	}

	f, ok := this.facilitys[fType]
	if ok {
		if int8(len(f.Levels)) >= level {
			 return f.Levels[level-1].Values
		}else{
			return []int{}
		}
	}else{
		return []int{}
	}
}


func (this *facilityConf) GetAdditions(fType int8) []int8 {
	f, ok := this.facilitys[fType]
	if ok {
		return f.Additions
	}else{
		return []int8{}
	}
}




