package model

type RoleRes struct {
	DB 				dbSync		`xorm:"-"`
	Id				int			`xorm:"id pk autoincr"`
	RId				int			`xorm:"rid"`
	Wood			int			`xorm:"wood"`
	Iron			int			`xorm:"iron"`
	Stone			int			`xorm:"stone"`
	Grain			int			`xorm:"grain"`
	Gold			int			`xorm:"gold"`
	Decree			int			`xorm:"decree"`	//令牌
	WoodYield		int			`xorm:"wood_yield"`
	IronYield		int			`xorm:"iron_yield"`
	StoneYield		int			`xorm:"stone_yield"`
	GrainYield		int			`xorm:"grain_yield"`
	GoldYield		int			`xorm:"gold_yield"`
	DepotCapacity	int			`xorm:"depot_capacity"`	//仓库容量
}

func (this *RoleRes) TableName() string {
	return "role_res"
}
