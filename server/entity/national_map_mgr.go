package entity

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"slgserver/config"
	"slgserver/log"
	"slgserver/model"
	"slgserver/util"
	"sort"
	"sync"
)

var MapWith = 40
var MapHeight = 40
const ScanWith = 3
const ScanHeight = 3

type NMArray struct {
	arr []model.NationalMap
}

func (this NMArray) Len() int {
	return len(this.arr)
}

func (this NMArray) Swap(i, j int) {
	this.arr[i], this.arr[j] = this.arr[j], this.arr[i]
}

func (this NMArray) Less(i, j int) bool {
	if this.arr[i].X < this.arr[j].X{
		return true
	}else if this.arr[i].X == this.arr[j].X {
		return this.arr[i].Y < this.arr[j].Y
	}else{
		return false
	}
}

type mapData struct {
	Width	int 			`json:"w"`
	Height	int				`json:"h"`
	List	[][]int			`json:"list"`
}

type NationalMapMgr struct {
	mutex sync.RWMutex
	conf map[int]model.NationalMap
	confArr NMArray
}

var NMMgr = &NationalMapMgr{
	conf: make(map[int]model.NationalMap),
}

func (this* NationalMapMgr) Load() {

	fileName := config.File.MustValue("logic", "map_data",
		"./data/conf/map.json")

	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.DefaultLog.Error("NationalMapMgr load file error", zap.Error(err))
		os.Exit(0)
	}

	m := &mapData{}
	err = json.Unmarshal(jdata, m)
	if err != nil {
		log.DefaultLog.Error("NationalMapMgr Unmarshal json error", zap.Error(err))
		os.Exit(0)
	}

	//转成服务用的结构
	MapWith = m.Width
	MapHeight = m.Height

	this.mutex.Lock()
	defer this.mutex.Unlock()

	for i, v := range m.List {
		t := v[0]
		l := v[1]
		this.conf[i] = model.NationalMap{X: i/MapWith, Y: i%MapWith, Id: i, Type: t, Level: l}
	}
	
	this.confArr.arr = make([]model.NationalMap, len(this.conf))
	i := 0
	for _, v := range this.conf {
		this.confArr.arr[i] = v
		i++
	}

	sort.Sort(this.confArr)

}

func (this* NationalMapMgr) Scan(x, y int) []model.NationalMap {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(40, x+ScanWith)

	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(40, y+ScanHeight)

	c := (maxX-minX+1)*(maxY-minY+1)
	r := make([]model.NationalMap, c)

	index := 0
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			r[index] = this.confArr.arr[i*ScanWith+j]
			index++
		}
	}
	return r
}