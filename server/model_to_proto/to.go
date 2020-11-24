package model_to_proto

import (
	"slgserver/model"
	"slgserver/server/proto"
)

func Role(m *model.Role, p *proto.Role)  {
	p.UId = m.UId
	p.SId = m.SId
	p.RId = m.RId
	p.Sex = m.Sex
	p.NickName = m.NickName
	p.HeadId = m.HeadId
	p.Balance = m.Balance
	p.Profile = m.Profile
}

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

func MRBuild(m *model.MapRoleBuild, p *proto.MapRoleBuild)  {
	p.X = m.X
	p.Y = m.Y
	p.Type = m.Type
	p.CurDurable = m.CurDurable
	p.MaxDurable = m.MaxDurable
	p.Level = m.Level
	p.RId = m.RId
	p.Name = m.Name
	p.Defender = m.Defender
}

func MCBuild(m *model.MapRoleCity, p *proto.MapRoleCity)  {
	p.X = m.X
	p.Y = m.Y
	p.CityId = m.CityId
	p.CurDurable = m.CurDurable
	p.MaxDurable = m.MaxDurable
	p.Level = m.Level
	p.RId = m.RId
	p.Name = m.Name
	p.IsMain = m.IsMain == 1
}


func Army(m *model.Army, p *proto.Army)  {
	p.CityId = m.CityId
	p.Id = m.Id
	p.Order = m.Order
	p.Generals = m.GeneralArray
	p.Soldiers = m.SoldierArray
	p.Cmd = m.Cmd
	p.State = m.State
	p.FromX = m.FromX
	p.FromY = m.FromY
	p.ToX = m.ToX
	p.ToY = m.ToY
	p.Start = m.Start.Unix()
	p.End = m.End.Unix()
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
