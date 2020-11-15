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

var FConf facilityConf

type conf struct {
	Name	string
	Type	int8
}

type facilityConf struct {
	Title	string	`json:"title"`
	List 	[]conf	`json:"list"`
}

func (this *facilityConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}

func (this *facilityConf) MaxLevel(fType int8) int8 {
	if t := FARMY.MaxLevel(fType); t != 0{
		return t
	}else if t = FBarrack.MaxLevel(fType); t != 0{
		return t
	}else if t = FCAMP.MaxLevel(fType); t != 0{
		return t
	}else if t = FFCT.MaxLevel(fType); t != 0{
		return t
	}else if t = FGEN.MaxLevel(fType); t != 0{
		return t
	}else if t = FMarket.MaxLevel(fType); t != 0{
		return t
	}else if t = FMBS.MaxLevel(fType); t != 0{
		return t
	}else if t = FPRC.MaxLevel(fType); t != 0{
		return t
	}else if t = FSJT.MaxLevel(fType); t != 0{
		return t
	}else if t = FWALL.MaxLevel(fType); t != 0{
		return t
	}
	return 0
}


