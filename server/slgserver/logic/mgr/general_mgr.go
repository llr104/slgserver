package mgr

import (
	"go.uber.org/zap"
	"math/rand"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf"
	"slgserver/server/slgserver/static_conf/general"
	"slgserver/util"
	"sync"
	"time"
)

type generalMgr struct {
	mutex     sync.RWMutex
	genByRole map[int][]*model.General
	genByGId  map[int]*model.General
}

var GMgr = &generalMgr{
	genByRole: make(map[int][]*model.General),
	genByGId: make(map[int]*model.General),
}

func (this*generalMgr) updatePhysicalPower() {
	limit := static_conf.Basic.General.PhysicalPowerLimit
	recoverCnt := static_conf.Basic.General.RecoveryPhysicalPower
	for true {
		time.Sleep(1*time.Hour)
		this.mutex.RLock()
		for _, g := range this.genByGId {
			if g.PhysicalPower < limit{
				g.PhysicalPower = util.MinInt(limit, g.PhysicalPower+recoverCnt)
				g.SyncExecute()
			}
		}
		this.mutex.RUnlock()
	}
}

func (this*generalMgr) createNPC() ([]*model.General, bool){
	//创建
	gs := make([]*model.General, 0)
	sess := db.MasterDB.NewSession()
	sess.Begin()

	for _, v := range general.General.GMap {
		if v.Star == 3{
			r, ok := this.NewGeneral(v.CfgId, 0)
			if ok == false {
				sess.Rollback()
				return nil, false
			}
			gs = append(gs, r)
		}

	}
	if err := sess.Commit(); err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}else{
		return gs, true
	}
}

func (this*generalMgr) add(g *model.General) {
	this.mutex.Lock()

	if _,ok := this.genByRole[g.RId]; ok == false{
		this.genByRole[g.RId] = make([]*model.General, 0)
	}
	this.genByRole[g.RId] = append(this.genByRole[g.RId], g)
	this.genByGId[g.Id] = g

	this.mutex.Unlock()
}



func (this*generalMgr) Load(){

	err := db.MasterDB.Table(model.General{}).Find(this.genByGId)
	if err != nil {
		log.DefaultLog.Warn("db error", zap.Error(err))
		return
	}

	for _, v := range this.genByGId {
		if _, ok := this.genByRole[v.RId]; ok==false {
			this.genByRole[v.RId] = make([]*model.General, 0)
		}
		this.genByRole[v.RId] = append(this.genByRole[v.RId], v)
	}

	if len(this.genByGId) == 0{
		this.createNPC()
	}

	go this.updatePhysicalPower()
}

func (this*generalMgr) GetByRId(rid int) ([]*model.General, bool){
	this.mutex.Lock()
	r, ok := this.genByRole[rid]
	this.mutex.Unlock()

	if ok {
		return r, true
	}

	gs := make([]*model.General, 0)
	err := db.MasterDB.Table(new(model.General)).Where("rid=?", rid).Find(&gs)
	if err == nil {
		if len(gs) > 0 {
			for _, g := range gs {
				this.add(g)
			}
			return gs, true
		}else{
			log.DefaultLog.Warn("general not fount", zap.Int("rid", rid))
			return nil, false
		}
	}else{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}
}

//查找将领
func (this*generalMgr) GetByGId(gid int) (*model.General, bool){
	this.mutex.RLock()
	g, ok := this.genByGId[gid]
	this.mutex.RUnlock()
	if ok {
		return g, true
	}else{

		g := &model.General{}
		r, err := db.MasterDB.Table(new(model.General)).Where("id=?", gid).Get(g)
		if err == nil{
			if r {
				this.add(g)
				return g, true
			}else{
				log.DefaultLog.Warn("general gid not found", zap.Int("gid", gid))
				return nil, false
			}

		}else{
			log.DefaultLog.Warn("general gid not found", zap.Int("gid", gid))
			return nil, false
		}
	}
}

//这个角色是否有这个武将
func (this*generalMgr) HasGeneral(rid int ,gid int) (*model.General,bool){
	r, ok := this.GetByRId(rid)
	if ok {
		for _, v := range r {
			t := v
			if t.Id == gid {
				return t,true
			}
		}
	}
	return nil,false
}

func (this*generalMgr) HasGenerals(rid int, gids []int) ([]*model.General,bool){
	gs := make([]*model.General, 0)
	for i := 0; i < len(gids); i++ {
		g,ok := this.HasGeneral(rid,gids[i])
		if ok{
			gs = append(gs,g)
		}else{
			return gs,false
		}
	}
	return gs,true
}

func (this*generalMgr) ActiveCount(rid int) int{
	gs, ok := this.GetByRId(rid)
	cnt := 0
	if ok {
		for _, g := range gs {
			if g.ParentId == 0{
				cnt += 1
			}
		}
	}
	return cnt
}

func (this*generalMgr) NewGeneral(cfgId int, rid int) (*model.General, bool) {
	cfg, ok := general.General.GMap[cfgId]
	if ok {
		g := &model.General{RId: rid, CfgId: cfg.CfgId, Order: 0, CityId: 0,
			PhysicalPower: static_conf.Basic.General.PhysicalPowerLimit,
			Level:         1, CreatedAt: time.Now(),CurArms: cfg.Arms[0],HasPrPoint: 0,UsePrPoint: 0,
			AttackDis: 0,ForceAdded: 0,StrategyAdded: 0,DefenseAdded: 0,SpeedAdded: 0,DestroyAdded: 0,
			Star: cfg.Star,StarLv: 0,ComposeType: 0,ParentId: 0,
		}

		if _, err := db.MasterDB.Table(model.General{}).Insert(g); err != nil {
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}else{
			this.add(g)
			return g, true
		}
	}else{
		return nil, false
	}
}

/*
如果不存在尝试去创建
*/
func (this*generalMgr) GetByRIdTryCreate(rid int) ([]*model.General, bool){
	r, ok := this.GetByRId(rid)
	if ok {
		return r, true
	}else{
		//创建
		gs := make([]*model.General, 0)
		sess := db.MasterDB.NewSession()
		sess.Begin()

		g, ok := this.RandCreateGeneral(rid,3)
		if ok == false{
			sess.Rollback()
			return nil, false
		}
		gs = g

		if err := sess.Commit(); err != nil{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}else{
			return gs, true
		}
	}
}


/*
随机创建一个
*/
func (this*generalMgr) RandCreateGeneral(rid int, nums int) ([]*model.General, bool){
	//创建
	gs := make([]*model.General, 0)
	sess := db.MasterDB.NewSession()
	sess.Begin()

	for i := 0; i < nums; i++ {
		r := rand.Intn(10) * 10
		d := this.PrToCfgId(r)
		g, ok := this.NewGeneral(d, rid)
		if ok == false{
			sess.Rollback()
			return nil, false
		}
		gs = append(gs, g)
	}

	if err := sess.Commit(); err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}else{
		return gs, true
	}
}


func (this*generalMgr) PrToCfgId(rate int) (cfgId int){
	gs := make([]int, 0)
	defgs := make([]int, 0)


	for i := 0;i < len(general.General.GArr);i++{
		if general.General.GArr[i].Probability >= 80{
			defgs = append(defgs, general.General.GArr[i].CfgId)
		}

	}

	for i := 0;i < len(general.General.GArr);i++{
		if general.General.GArr[i].Probability >= rate{
			gs = append(gs, general.General.GArr[i].CfgId)
		}

	}


	if len(gs) == 0{
		return defgs[0]
	}
	r := rand.Intn(len(gs))
	return gs[r]
}

//获取npc武将
func (this*generalMgr) GetNPCGenerals(cnt int) ([]model.General, bool) {
	gs, ok := this.GetByRId(0)
	if ok == false {
		return make([]model.General, 0), false
	}else{
		if cnt > len(gs){
			return make([]model.General, 0), false
		}else{
			m := make(map[int]int)
			for true {
				 r := rand.Intn(len(gs))
				 m[r] = r
				 if len(m) == cnt{
				 	break
				 }
			}
			rgs := make([]model.General, 0)
			for _, v := range m {
				t := *gs[v]
				rgs = append(rgs, t)
			}
			return rgs, true
		}
	}
}

func (this *generalMgr) GetDestroy(army *model.Army) int{
	destroy := 0
	for _, g := range army.Gens {
		if g != nil {
			destroy += g.GetDestroy()
		}
	}
	return destroy
}

//体力是否足够
func (this*generalMgr) PhysicalPowerIsEnough(army *model.Army, cost int) bool{
	for _, g := range army.Gens {
		if g == nil{
			continue
		}
		if g.PhysicalPower < cost{
			return false
		}
	}
	return true
}

//尝试使用体力
func (this *generalMgr) TryUsePhysicalPower(army *model.Army, cost int) bool{

	if this.PhysicalPowerIsEnough(army, cost) == false{
		return false
	}

	for _, g := range army.Gens {
		if g == nil{
			continue
		}
		g.PhysicalPower -= cost
		g.SyncExecute()
	}

	return true
}
