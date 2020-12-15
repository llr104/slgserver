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
var Basic basic



type conscript struct {
	Des       string `json:"des"`
	CostWood  int    `json:"cost_wood"`
	CostIron  int    `json:"cost_iron"`
	CostStone int    `json:"cost_stone"`
	CostGrain int    `json:"cost_grain"`
	CostGold  int    `json:"cost_gold"`
}

type general struct {
	Des                   string `json:"des"`
	PhysicalPowerLimit    int    `json:"physical_power_limit"`    //体力上限
	CostPhysicalPower     int    `json:"cost_physical_power"`     //消耗体力
	RecoveryPhysicalPower int    `json:"recovery_physical_power"` //恢复体力
	ReclamationTime       int    `json:"reclamation_time"`        //屯田消耗时间，单位秒
	ReclamationCost       int    `json:"reclamation_cost"`        //屯田消耗政令
	DrawGeneralCost       int    `json:"draw_general_cost"`        //抽卡消耗金币
	PrPoint               int    `json:"pr_pont"`                  //合成一个武将或者的技能点

}

type role struct {
	Des           string 	`json:"des"`
	Wood          int 		`json:"wood"`
	Iron          int 		`json:"iron"`
	Stone         int 		`json:"stone"`
	Grain         int 		`json:"grain"`
	Gold          int 		`json:"gold"`
	Decree        int 		`json:"decree"`
	WoodYield     int 		`json:"wood_yield"`
	IronYield     int 		`json:"iron_yield"`
	StoneYield    int 		`json:"stone_yield"`
	GrainYield    int 		`json:"grain_yield"`
	GoldYield     int 		`json:"gold_yield"`
	DepotCapacity int 		`json:"depot_capacity"`
	BuildLimit    int 		`json:"build_limit"`
	RecoveryTime  int 		`json:"recovery_time"`
	DecreeLimit   int 		`json:"decree_limit"`
}

type city struct {
	Des           string `json:"des"`
	Cost          int8   `json:"cost"`
	Durable       int    `json:"durable"`
	RecoveryTime  int    `json:"recovery_time"`
	TransformRate int    `json:"transform_rate"`
}

type npcLevel struct {
	Soilders int `json:"soilders"`
}

type npc struct {
	Des    string     `json:"des"`
	Levels []npcLevel `json:"levels"`
}

type union struct {
	Des         string `json:"des"`
	MemberLimit int    `json:"member_limit"`
}

type basic struct {
	ConScript conscript `json:"conscript"`
	General   general   `json:"general"`
	Role      role      `json:"role"`
	City      city      `json:"city"`
	Npc       npc       `json:"npc"`
	Union     union     `json:"union"`
}

func (this *basic) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "basic.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("basic load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	fmt.Println(this)
}

func (this *basic) GetNPC(level int8) (*npcLevel, bool){
	if level <= 0{
		return nil, false
	}
	if len(this.Npc.Levels) >= int(level){
		return &this.Npc.Levels[level-1], true
	}
	return nil, false
}
