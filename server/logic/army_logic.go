package logic

import (
	"slgserver/model"
	"slgserver/server"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
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
					AMgr.PushAction(army)
					//战斗还要加

				}else if cur_t >= 1*diff + army.Start.Unix(){
					//一倍路程
					army.State = model.ArmyBack

					//战斗还要加
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
