package logic

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"math"
	"os"
	"slgserver/config"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/global"
	"slgserver/server/model"
	"slgserver/util"
	"sync"
)


const ScanWith = 3
const ScanHeight = 3


type NMArray struct {
	arr []model.NationalMap
}


type mapData struct {
	Width	int 			`json:"w"`
	Height	int				`json:"h"`
	List	[][]int			`json:"list"`
}

func ToPosition(x, y int) int {
	return x+global.MapHeight*y
}

func Distance(begX, begY, endX, endY int) float64 {
	w := math.Abs(float64(endX - begX))
	h := math.Abs(float64(endY - begY))
	return math.Sqrt(w*w + h*h)
}

func TravelTime(speed, begX, begY, endX, endY int) int {
	dis := Distance(begX, begY, endX, endY)
	t := dis / float64(speed)*1000000
	return int(t)
}

type NationalMapMgr struct {
	mutex sync.RWMutex
	conf map[int]model.NationalMap
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
	global.MapWith = m.Width
	global.MapHeight = m.Height

	this.mutex.Lock()
	defer this.mutex.Unlock()

	temp := make([]model.NationalMap, 0)
	isNeedDb := false
	cnt, err := db.MasterDB.Table(new(model.NationalMap)).Count(&temp)
	if cnt == 0 && err == nil{
		isNeedDb = true
	}

	for i, v := range m.List {
		t := int8(v[0])
		l := int8(v[1])
		d := model.NationalMap{X: i/global.MapWith, Y: i%global.MapWith, MId: i, Type: t, Level: l}
		this.conf[i] = d
		if isNeedDb {
			db.MasterDB.Insert(d)
		}
	}

}

func (this* NationalMapMgr) IsCanBuild(x, y int) bool {
	posIndex := ToPosition(x, y)
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	c,ok := this.conf[posIndex]
	if ok {
		if c.Type == 0{
			return false
		}else {
			return true
		}
	}else {
		return false
	}
}

func (this* NationalMapMgr) PositionBuild(x, y int) (model.NationalMap, bool) {
	posIndex := ToPosition(x, y)
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	b, ok := this.conf[posIndex]
	return b, ok
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
			v, ok := this.conf[ToPosition(i, j)]
			if ok {
				r[index] = v
			}
			index++
		}
	}
	return r
}