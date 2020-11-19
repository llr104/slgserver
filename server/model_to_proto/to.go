package model_to_proto

import (
	"slgserver/model"
	"slgserver/server/proto"
)

func RRes(m *model.RoleRes, p *proto.RoleRes) {
	p.Gold = m.Gold
	p.Grain = m.Grain
	p.Stone = m.Stone
	p.Iron = m.Iron
	p.Wood = m.Wood
	p.Decree = m.Decree
	p.GoldYield = m.GoldYield
	p.GrainYield = m.GrainYield
	p.StoneYield = m.StoneYield
	p.IronYield = m.IronYield
	p.WoodYield = m.WoodYield
	p.DepotCapacity = m.DepotCapacity
}

func Army(m *model.Army, p *proto.Army)  {
	p.CityId = m.CityId
	p.Id = m.Id
	p.Order = m.Order
	p.FirstId = m.FirstId
	p.SecondId = m.SecondId
	p.ThirdId = m.ThirdId
	p.FirstSoldierCnt = m.FirstSoldierCnt
	p.SecondSoldierCnt = m.SecondSoldierCnt
	p.ThirdSoldierCnt = m.ThirdSoldierCnt
}

func General(m *model.General, p *proto.General)  {
	p.CityId = m.CityId
	p.Order = m.Order
	p.Cost = m.Cost
	p.Speed = m.Speed
	p.Defense = m.Defense
	p.Strategy = m.Strategy
	p.Force = m.Force
	p.Name = m.Name
	p.Id = m.Id
	p.CfgId = m.CfgId
	p.Destroy = m.Destroy
	p.Level = m.Level
	p.Exp = m.Exp
}
