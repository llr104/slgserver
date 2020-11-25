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



//地图资源配置
var MapBuildConf mapBuildConf


type cfg struct {
	Type     int8   `json:"type"`
	Name     string `json:"name"`
	Level    int8   `json:"level"`
	Grain    int    `json:"grain"`
	Wood     int    `json:"wood"`
	Iron     int    `json:"iron"`
	Stone    int    `json:"stone"`
	Durable  int    `json:"durable"`
	Defender int    `json:"defender"`
}


type mapBuildConf struct {
	Title   string 	`json:"title"`
	Cfg		[]cfg	`json:"cfg"`
}

func (this *mapBuildConf) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "map_build.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("mapBuildConf load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	fmt.Println(this)
}

func (this* mapBuildConf) BuildConfig(cfgType int8, level int8) (*cfg, bool) {
	for _, v := range this.Cfg {
		if v.Level == level && v.Type == cfgType{
			return &v, true
		}
	}
	return nil, false
}

