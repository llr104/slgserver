package mgr

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"math"
	"os"
	"slgserver/config"
	"slgserver/log"
	"slgserver/server/slgserver/global"
	"slgserver/server/slgserver/model"
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


func Distance(begX, begY, endX, endY int) float64 {
	w := math.Abs(float64(endX - begX))
	h := math.Abs(float64(endY - begY))
	return math.Sqrt(w*w + h*h)
}

func TravelTime(speed, begX, begY, endX, endY int) int {
	dis := Distance(begX, begY, endX, endY)
	t := dis / float64(speed)*100000000
	return int(t)
}

type NationalMapMgr struct {
	mutex sync.RWMutex
	conf map[int]model.NationalMap
	sysCity map[int]model.NationalMap
}

var NMMgr = &NationalMapMgr{
	conf: make(map[int]model.NationalMap),
	sysCity: make(map[int]model.NationalMap),
}

func (this*NationalMapMgr) Load() {

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

	for i, v := range m.List {
		t := int8(v[0])
		l := int8(v[1])
		d := model.NationalMap{Y: i/ global.MapHeight, X: i% global.MapWith, MId: i, Type: t, Level: l}
		this.conf[i] = d
		if d.Type == model.MapBuildSysCity{
			this.sysCity[i] = d
		}
	}

	fmt.Println(this.sysCity)

}

func (this*NationalMapMgr) IsCanBuild(x, y int) bool {
	posIndex := global.ToPosition(x, y)
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

func (this*NationalMapMgr) IsCanBuildCity(x, y int) bool {

	//系统城池附近5格不能有玩家城池
	for _, nationalMap := range this.sysCity {
		if x >= nationalMap.X - 5 && x <= nationalMap.X + 5 &&
			y >= nationalMap.Y - 5 && y <= nationalMap.Y + 5{
			return false
		}
	}

	for i := x-2; i <= x+2; i++ {
		if i < 0 || i > global.MapWith {
			return false
		}

		for j := y-2; j <= y+2; j++ {
			if j < 0 || j > global.MapHeight {
				return false
			}
		}

		if this.IsCanBuild(x, y) == false ||
			RBMgr.IsEmpty(x, y) == false ||
			RCMgr.IsEmpty(x, y) == false{
			return false
		}
	}
	return true
}

func (this*NationalMapMgr) MapResTypeLevel(x, y int) (bool, int8, int8) {
	n, ok := this.PositionBuild(x, y)
	if ok {
		return true, n.Type, n.Level
	}
	return false, 0, 0
}

func (this*NationalMapMgr) PositionBuild(x, y int) (model.NationalMap, bool) {
	posIndex := global.ToPosition(x, y)
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	b, ok := this.conf[posIndex]
	return b, ok
}

func (this*NationalMapMgr) Scan(x, y int) []model.NationalMap {
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
			v, ok := this.conf[global.ToPosition(i, j)]
			if ok {
				r[index] = v
			}
			index++
		}
	}
	return r
}