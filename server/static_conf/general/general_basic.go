package general

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"slgserver/config"
	"slgserver/log"
)

var GenBasic Basic

type gLevel struct {
	Level		int8`json:"level"`
	Exp			int `json:"exp"`
	Soldiers	int `json:"soldiers"`
}

type Basic struct {
	Title	string		`json:"title"`
	Levels	[]gLevel	`json:"levels"`
}

func (this *Basic) Load()  {
	jsonDir := config.File.MustValue("logic", "json_data", "../data/conf/")
	fileName := path.Join(jsonDir, "general", "general_basic.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("general load file error", zap.Error(err), zap.String("file", fileName))
		os.Exit(0)
	}

	json.Unmarshal(jdata, this)

	fmt.Println(this)

	General.Load()
}

func (this *Basic) GetLevel(l int8) (*gLevel, error){
	if l <= 0{
		return nil, errors.New("level error")
	}
	if int(l) < len(this.Levels){
		return &this.Levels[l-1], nil
	}else{
		return nil, errors.New("level error")
	}
}

