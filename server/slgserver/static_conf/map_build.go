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
	Title   string   `json:"title"`
	Cfg		[]cfg `json:"cfg"`
	cfgMap  map[int8][]cfg
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

	this.cfgMap = make(map[int8][]cfg)
	for _, v := range this.Cfg {
		if _, ok := this.cfgMap[v.Type]; ok == false{
			this.cfgMap[v.Type] = make([]cfg, 0)
		}
		this.cfgMap[v.Type] = append(this.cfgMap[v.Type], v)
	}
}

func (this*mapBuildConf) BuildConfig(cfgType int8, level int8) (*cfg, bool) {
	if c, ok := this.cfgMap[cfgType]; ok {
		for _, v := range c {
			if v.Level == level {
				return &v, true
			}
		}
	}

	return nil, false
}

