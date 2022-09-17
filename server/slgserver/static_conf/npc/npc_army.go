package npc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"

	"github.com/llr104/slgserver/config"
	"github.com/llr104/slgserver/log"
	"go.uber.org/zap"
)

var Cfg npc

type ArmyCfg struct {
	Lvs    []int8 `json:"lvs"`
	CfgIds []int  `json:"cfgIds"`
}

type armyArray struct {
	Des      string    `json:"des"`
	Soldiers int       `json:"soldiers"`
	ArmyCfg  []ArmyCfg `json:"army"`
}

type npc struct {
	Des   string      `json:"des"`
	Armys []armyArray `json:"armys"`
}

func (this *npc) Load() {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "npc", "npc_army.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("npc_army load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	fmt.Println(this)
}

func (this *npc) NPCSoilder(level int8) int {
	if int(level) > len(this.Armys) || level <= 0 {
		return 0
	}
	return this.Armys[level-1].Soldiers
}

func (this *npc) RandomOne(level int8) (bool, *ArmyCfg) {
	if int(level) > len(this.Armys) || level <= 0 {
		return false, nil
	}

	r := rand.Intn(len(this.Armys[level-1].ArmyCfg))
	return true, &this.Armys[level-1].ArmyCfg[r]
}
