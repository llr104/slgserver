package logic

import (
	"slgserver/model"
	"sync"
)

var SysArmy* sysArmyLogic

func init() {
	SysArmy = &sysArmyLogic{
		sysArmys: make(map[int]*[]model.Army),
	}
}

type sysArmyLogic struct {
	mutex 		sync.Mutex
	sysArmys    map[int]*[]model.Army   //key:posId 系统建筑军队
}



