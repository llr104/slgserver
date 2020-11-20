package logic

import (
	"go.uber.org/zap"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"slgserver/util"
	"time"
)


var ArmyLogic *armyLogic

func init() {
	ArmyLogic = &armyLogic{armys: make(chan *model.Army, 100)}
	go ArmyLogic.running()
}

type armyLogic struct {
	armys    chan *model.Army
}

func (this *armyLogic) running(){
	for {
		select {
		case army := <-this.armys:
			cur_t := time.Now().Unix()
			if army.State == model.ArmyAttack {
				diff := army.End.Unix() - army.Start.Unix()

				if cur_t >= 2*diff + army.Start.Unix() {
					//两倍路程
					army.State = model.ArmyIdle
					army.ToX = army.FromX
					army.ToY = army.FromY
					this.battle(army)
					AMgr.PushAction(army)
				}else if cur_t >= 1*diff + army.Start.Unix(){
					//一倍路程
					army.State = model.ArmyBack
					this.battle(army)
				}
			}

			army.NeedUpdate = true
			ap := &proto.ArmyStatePush{}
			ap.CityId = army.CityId
			model_to_proto.Army(army, &ap.Army)
			//通知部队变化了
			server.DefaultConnMgr.PushByRoleId(army.RId, "general.armyState", ap)
		}
	}
}

func (this *armyLogic) Arrive(army* model.Army) {
	this.armys <- army
}

//简单战斗
func (this *armyLogic) battle(army* model.Army) {
	_, ok := RCMgr.PositionCity(army.ToX, army.ToY)
	if ok {
		//打城池
		army.FirstSoldierCnt = 0
		army.SecondSoldierCnt = 0
		army.ThirdSoldierCnt = 0
		return
	}

	_, ok = RBMgr.PositionBuild(army.ToX, army.ToY)
	if ok {
		//打玩家占领的领地
		army.FirstSoldierCnt = util.MinInt(0, army.FirstSoldierCnt-50)
		army.SecondSoldierCnt = util.MinInt(0, army.SecondSoldierCnt-50)
		army.ThirdSoldierCnt = util.MinInt(0, army.ThirdSoldierCnt-50)
	}else{
		army.FirstSoldierCnt = util.MinInt(0, army.FirstSoldierCnt-10)
		army.SecondSoldierCnt = util.MinInt(0, army.SecondSoldierCnt-10)
		army.ThirdSoldierCnt = util.MinInt(0, army.ThirdSoldierCnt-10)
	}

	//占领
	this.OccupyBuild(army.Id, army.ToX, army.ToY)


}

func (this* armyLogic) OccupyBuild(rid, x, y int)  {
	newId := rid

	b, ok := RBMgr.PositionBuild(x, y)
	if ok {

		oldId := b.RId
		log.DefaultLog.Info("battle in role build",
			zap.Int("oldRId", oldId),
			zap.Int("newRId", newId))

		//被占领的减产
		if oldRole, ok := RResMgr.Get(oldId); ok{
			oldRole.WoodYield -= b.Wood
			oldRole.GrainYield -= b.Grain
			oldRole.StoneYield -= b.Stone
			oldRole.IronYield -= b.Iron
			oldRole.NeedUpdate = true
		}
		//占领的增加产量
		if newRole, ok := RResMgr.Get(newId); ok{
			newRole.WoodYield -= b.Wood
			newRole.GrainYield -= b.Grain
			newRole.StoneYield -= b.Stone
			newRole.IronYield -= b.Iron
			newRole.NeedUpdate = true
		}
		b.NeedUpdate = true
		b.RId = rid

		push := &proto.RoleBuildStatePush{}
		model_to_proto.MRBuild(b, &push.MRBuild)
		server.DefaultConnMgr.PushByRoleId(oldId, "role.roleBuildState", push)
		server.DefaultConnMgr.PushByRoleId(newId, "role.roleBuildState", push)

	}else{
		if NMMgr.IsCanBuild(x, y){
			b, ok := NMMgr.PositionBuild(x, y)
			if ok {
				if cfg, _ := BCMgr.BuildConfig(b.Type, b.Level); cfg != nil{
					rb := &model.MapRoleBuild{RId: rid, X: x, Y: y,
						Type: b.Type, Level: b.Level, Name: cfg.Name,
						Wood: cfg.Wood, Iron: cfg.Iron, Stone: cfg.Stone,
						Grain: cfg.Grain, CurDurable: cfg.Durable,
						MaxDurable: cfg.Durable, NeedUpdate: true}

					RBMgr.AddBuild(rb)

					//占领的增加产量
					if newRole, ok := RResMgr.Get(newId); ok{
						newRole.WoodYield -= rb.Wood
						newRole.GrainYield -= rb.Grain
						newRole.StoneYield -= rb.Stone
						newRole.IronYield -= rb.Iron
						newRole.NeedUpdate = true
					}

					push := &proto.RoleBuildStatePush{}
					model_to_proto.MRBuild(rb, &push.MRBuild)
					server.DefaultConnMgr.PushByRoleId(newId, "role.roleBuildState", push)
				}
			}
		}
	}
}