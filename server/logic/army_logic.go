package logic

import (
	"fmt"
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
			diff := army.End.Unix() - army.Start.Unix()
			if army.State == model.ArmyAttack {

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
					army.Start = army.End
					army.End = army.Start.Add(time.Duration(diff)*time.Second)

					fmt.Println(army.Start.Unix(), army.End.Unix(), army.End.Unix()-army.Start.Unix())
					this.battle(army)
					AMgr.PushAction(army)
				}
			}else if army.State == model.ArmyBack {
				 if cur_t >= 1*diff + army.Start.Unix(){
					//一倍路程
					army.State = model.ArmyIdle
				}
			}

			army.DB.Sync()
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
		return
	}

	_, ok = RBMgr.PositionBuild(army.ToX, army.ToY)
	if ok {
		//打玩家占领的领地
		army.SoldierArray[0] = util.MaxInt(0, army.SoldierArray[0]-50)
		army.SoldierArray[1] = util.MaxInt(0, army.SoldierArray[1]-50)
		army.SoldierArray[2] = util.MaxInt(0, army.SoldierArray[2]-50)
		this.OccupyRoleBuild(army.RId, army.ToX, army.ToY)
	}else{
		army.SoldierArray[0] = util.MaxInt(0, army.SoldierArray[0]-10)
		army.SoldierArray[1] = util.MaxInt(0, army.SoldierArray[1]-10)
		army.SoldierArray[2] = util.MaxInt(0, army.SoldierArray[2]-10)

		//占领
		this.OccupySystemBuild(army.RId, army.ToX, army.ToY)
	}

}

func (this* armyLogic) OccupyRoleBuild(rid, x, y int)  {
	newId := rid

	if b, ok := RBMgr.PositionBuild(x, y); ok {

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
			oldRole.DB.Sync()
		}
		//占领的增加产量
		if newRole, ok := RResMgr.Get(newId); ok{
			newRole.WoodYield += b.Wood
			newRole.GrainYield += b.Grain
			newRole.StoneYield += b.Stone
			newRole.IronYield += b.Iron
			newRole.DB.Sync()
		}
		b.DB.Sync()
		b.RId = rid

		push := &proto.RoleBuildStatePush{}
		model_to_proto.MRBuild(b, &push.MRBuild)
		server.DefaultConnMgr.PushByRoleId(oldId, "role.roleBuildState", push)
		server.DefaultConnMgr.PushByRoleId(newId, "role.roleBuildState", push)
	}
}

func (this* armyLogic) OccupySystemBuild(rid, x, y int)  {
	newId := rid

	if _, ok := RBMgr.PositionBuild(x, y); ok {
		return
	}

	if NMMgr.IsCanBuild(x, y){
		if b, ok := NMMgr.PositionBuild(x, y); ok {
			if cfg, _ := BCMgr.BuildConfig(b.Type, b.Level); cfg != nil{
				rb := &model.MapRoleBuild{RId: rid, X: x, Y: y,
					Type: b.Type, Level: b.Level, Name: cfg.Name,
					Wood: cfg.Wood, Iron: cfg.Iron, Stone: cfg.Stone,
					Grain: cfg.Grain, CurDurable: cfg.Durable,
					MaxDurable: cfg.Durable}
				RBMgr.AddBuild(rb)
				//占领的增加产量
				if newRole, ok := RResMgr.Get(newId); ok{
					newRole.WoodYield += rb.Wood
					newRole.GrainYield += rb.Grain
					newRole.StoneYield += rb.Stone
					newRole.IronYield += rb.Iron
					newRole.DB.Sync()
				}
				push := &proto.RoleBuildStatePush{}
				model_to_proto.MRBuild(rb, &push.MRBuild)
				server.DefaultConnMgr.PushByRoleId(newId, "role.roleBuildState", push)
			}
		}
	}
}