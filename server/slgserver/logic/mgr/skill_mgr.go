package mgr

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/model"
	"sync"
)

type skillMgr struct {
	mutex  sync.RWMutex
	skillMap map[int][]*model.Skill
}

var SkillMgr = &skillMgr{
	skillMap: make(map[int][]*model.Skill),
}


func (this*skillMgr) Load() {

	rr := make([]*model.Skill, 0)
	err := db.MasterDB.Find(&rr)
	if err != nil {
		log.DefaultLog.Error("skillMgr load role_res table error")
	}

	for _, v := range rr {
		if this.skillMap[v.RId] == nil{
			this.skillMap[v.RId] = make([]*model.Skill, 0)
		}
		this.skillMap[v.RId] = append(this.skillMap[v.RId], v)
	}
}


func (this*skillMgr) Get(rid int) ([]*model.Skill, bool){

	this.mutex.RLock()
	r, ok := this.skillMap[rid]
	this.mutex.RUnlock()

	if ok {
		return r, true
	}

	m := make([]*model.Skill, 0)
	ok, err := db.MasterDB.Table(new(model.Skill)).Where("rid=?", rid).Get(&m)
	if ok {

		this.mutex.Lock()
		this.skillMap[rid] = m
		this.mutex.Unlock()

		return m, true
	}else{
		if err == nil{
			log.DefaultLog.Warn("skill not found", zap.Int("rid", rid))
			return nil, false
		}else{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}
	}
}

func (this*skillMgr) GetSkillOrCreate(rid int, cfg int) (*model.Skill, bool){

	success := true
	m, ok := this.Get(rid)
	var ret *model.Skill = nil
	if ok {
		for _, v := range m {
			if v.CfgId != cfg{
				continue
			}else{
				ret = v
			}
		}
	}

	if ret == nil {
		ret = model.NewSkill(rid, cfg)
		_, err := db.MasterDB.InsertOne(ret)
		if err != nil {
			log.DefaultLog.Warn("db error", zap.Error(err))
			success = false
		}else{
			if this.skillMap[rid] == nil{
				this.skillMap[rid] = make([]*model.Skill, 0)
			}
			this.skillMap[rid] = append(this.skillMap[rid], ret)
		}
	}
	return ret, success

}
