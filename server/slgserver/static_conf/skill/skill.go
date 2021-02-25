package skill

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

var Skill skill

type skill struct {
	skills []Conf
	skillConfMap map[int]Conf
	outline outline
}


func (this *skill) Load()  {
	this.skills = make([]Conf, 0)
	this.skillConfMap = make(map[int]Conf)

	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "skill", "skill_outline.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("skill load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, &this.outline)

	rd, err := ioutil.ReadDir(path.Join(jsonDir, "skill"))
	if err != nil {
		log.DefaultLog.Error("skill readdir error", zap.Error(err))
		os.Exit(0)
	}

	for _, r := range rd {
		if r.IsDir() {
			this.readSkill(path.Join(jsonDir, "skill", r.Name()))
		}
	}


	fmt.Println(this)
}

func (this *skill) readSkill(dir string) {
	rd, err := ioutil.ReadDir(dir)
	if err == nil{
		for _, r := range rd {
			if r.IsDir() == false{
				jdata, err := ioutil.ReadFile(path.Join(dir, r.Name()))
				if err == nil {
					conf := Conf{}
					if err := json.Unmarshal(jdata, &conf); err == nil{
						this.skills = append(this.skills, conf)
						this.skillConfMap[conf.CfgId] = conf
					}else{
						log.DefaultLog.Warn("Unmarshal skill error", zap.Error(err),
							zap.String("file", path.Join(dir, r.Name())))
					}
				}
			}
		}
	}

}

func (this *skill) GetCfg(cfgId int) (Conf, bool) {
	cfg, ok := this.skillConfMap[cfgId]
	return cfg, ok
}