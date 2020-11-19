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
