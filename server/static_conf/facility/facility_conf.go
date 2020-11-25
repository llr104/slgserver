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

type NeedRes struct {
	Decree 		int	`json:"decree"`
	Grain		int `json:"grain"`
	Wood		int `json:"wood"`
	Iron		int `json:"iron"`
	Stone		int `json:"stone"`
	Gold		int	`json:"gold"`
}

type iFacility interface {
	MaxLevel(fType int8) int8
	Need(fType int8, level int) (*NeedRes, bool)
	IsContain(t int8) bool
}
var FConf facilityConf

type conf struct {
	Name	string
	Type	int8
}

type facilityConf struct {
	Title	string `json:"title"`
	List 	[]conf  `json:"list"`
	loaders	[]iFacility
}

func (this *facilityConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "facility", "facility.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("facilityConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)

	FGEN.Load()
	FPRC.Load()
	FMBS.Load()
	FARMY.Load()
	FCAMP.Load()
	FBarrack.Load()
	FFCT.Load()
	FWALL.Load()
	FMarket.Load()
	FSJT.Load()
	FWareHouse.Load()

	this.loaders = make([]iFacility, 0)
	this.loaders = append(this.loaders, &FGEN)
	this.loaders = append(this.loaders, &FPRC)
	this.loaders = append(this.loaders, &FMBS)
	this.loaders = append(this.loaders, &FARMY)
	this.loaders = append(this.loaders, &FCAMP)
	this.loaders = append(this.loaders, &FBarrack)
	this.loaders = append(this.loaders, &FFCT)
	this.loaders = append(this.loaders, &FWALL)
	this.loaders = append(this.loaders, &FMarket)
	this.loaders = append(this.loaders, &FSJT)
	this.loaders = append(this.loaders, &FWareHouse)

}

func (this *facilityConf) MaxLevel(fType int8) int8 {
	for _, v := range this.loaders {
		if v.IsContain(fType){
			return v.MaxLevel(fType)
		}
	}
	return 0
}

func (this *facilityConf) Need(fType int8, level int) (*NeedRes, bool) {
	if level <= 0{
		return nil, false
	}

	for _, v := range this.loaders {
		if v.IsContain(fType){
			return v.Need(fType, level)
		}
	}

	log.DefaultLog.Info("facilityConf type not found", zap.Int("type", int(fType)))
	return nil, false
}





