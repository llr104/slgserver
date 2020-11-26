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

func MRBuild(m *model.MapRoleBuild, p *proto.MapRoleBuild, rNick string)  {
	p.RNick = rNick
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
	p.Destroy = m.Destroy
	p.Force = m.Force
	p.SpeedGrow = m.SpeedGrow
	p.DefenseGrow = m.DefenseGrow
	p.StrategyGrow = m.StrategyGrow
	p.DestroyGrow = m.DestroyGrow
	p.ForceGrow = m.ForceGrow
	p.PhysicalPower = m.PhysicalPower
	p.Name = m.Name
	p.Id = m.Id
	p.CfgId = m.CfgId
	p.Level = m.Level
	p.Exp = m.Exp
}

func WarReport(m *model.WarReport, p *proto.WarReport)  {
	p.CTime = m.CTime.UnixNano()/1e6
	p.Id = m.Id
	p.AttackRid = m.AttackRid
	p.DefenseRid = m.DefenseRid
	p.BegAttackArmy = m.BegAttackArmy
	p.BegDefenseArmy = m.BegDefenseArmy
	p.EndAttackArmy = m.EndAttackArmy
	p.EndDefenseArmy = m.EndDefenseArmy
	p.BegAttackGeneral = m.BegAttackGeneral
	p.BegDefenseGeneral = m.BegDefenseGeneral
	p.EndAttackGeneral = m.EndAttackGeneral
	p.EndDefenseGeneral = m.EndDefenseGeneral
	p.AttackIsWin = m.AttackIsWin
	p.AttackIsRead = m.AttackIsRead
	p.DefenseIsRead = m.DefenseIsRead
	p.DestroyDurable = m.DestroyDurable
	p.Occupy = m.Occupy
	p.X = m.X
	p.X = m.X

}
