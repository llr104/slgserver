package general

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

var General general

type g struct {
	Name     		string	`json:"name"`
	CfgId    		int		`json:"cfgId"`
	Force    		int		`json:"force"`
	Strategy 		int		`json:"strategy"`
	Defense  		int		`json:"defense"`
	Speed    		int		`json:"speed"`
	Destroy      	int   	`json:"destroy"`
	ForceGrow    	int   	`json:"force_grow"`
	StrategyGrow 	int   	`json:"strategy_grow"`
	DefenseGrow  	int   	`json:"defense_grow"`
	SpeedGrow   	int   	`json:"speed_grow"`
	DestroyGrow  	int   	`json:"destroy_grow"`
	Cost         	int8  	`json:"cost"`
	Probability  	int   	`json:"probability"`
	Star         	int8   	`json:"star"`
	Arms         	[]int 	`json:"arms"`
	Camp         	int8  	`json:"camp"`
}

type general struct {
	Title string `json:"title"`
	GArr  []g    `json:"list"`
	GMap  map[int]g
}

func (this *general) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "general", "general.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("general load file error",
			zap.Error(err),
			zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)
	this.GMap = make(map[int]g)
	for _, v := range this.GArr {
		this.GMap[v.CfgId] = v
	}
	fmt.Println(this)
}

func (this *general) Cost(cfgId int) int8 {
	c, ok := this.GMap[cfgId]
	if ok {
		return c.Cost
	}else{
		return 0
	}

}