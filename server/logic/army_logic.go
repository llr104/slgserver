package logic

import (
	"slgserver/model"
	"slgserver/server"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"time"
)

type ArmyLogic struct {

}

func (this *ArmyLogic) Arrive(army* model.Army) {
	//先让他原路返回
	if army.State != model.ArmyBack {
		diff := army.End.Unix() - army.Start.Unix()
		army.Start = army.End
		army.End = army.Start.Add(time.Duration(diff))
		army.State = model.ArmyBack
	}else{
		army.ToX = army.FromX
		army.ToY = army.FromY
		army.State = model.ArmyIdle
	}
	army.NeedUpdate = true

	ap := &proto.ArmyStatePush{}
	ap.CityId = army.CityId
	model_to_proto.Army(army, &ap.Army)
	//通知部队变化了
	server.DefaultConnMgr.PushByRoleId(army.RId, "general.armyState", ap)
}
