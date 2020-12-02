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

var GenArms Arms


type gArmsCondit struct {
	Level		    int     `json:"level"`
	StarLevel		int     `json:"star_lv"`
}


type gArmsCost struct {
	Gold		    int     `json:"gold"`
}


type gArms struct {
	Id		    int             `json:"id"`
	Name		string          `json:"name"`
	Condition   gArmsCondit     `json:"condition"`
	ChangeCost  gArmsCost       `json:"change_cost"`
	HarmRatio	[][]int			`json:"harm_ratio"`
}


type Arms struct {
	Title	string		 `json:"title"`
	Arms	[]gArms	     `json:"arms"`
	AMap    map[int]gArms
}


func (this *Arms) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "general", "general_arms.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("general load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	this.AMap = make(map[int]gArms)
	for _, v := range this.Arms {
		this.AMap[v.Id] = v
	}

	fmt.Println(this)
}

func (this *Arms) GetArm(id int) (gArms, error){
	return this.AMap[id], nil
}




