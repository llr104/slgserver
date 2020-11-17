package model

type RoleRes struct {
	Id				int			`json:"id" xorm:"id pk autoincr"`
	RId				int			`json:"uid" xorm:"rid"`
	Wood			int			`json:"wood"`
	Iron			int			`json:"iron"`
	Stone			int			`json:"stone"`
	Grain			int			`json:"grain"`
	Gold			int			`json:"gold"`
	Decree			int			`json:"decree"`	//令牌
	WoodYield		int			`json:"wood_yield"`
	IronYield		int			`json:"iron_yield"`
	StoneYield		int			`json:"stone_yield"`
	GrainYield		int			`json:"grain_yield"`
	GoldYield		int			`json:"gold_yield"`
	DepotCapacity	int			`json:"depot_capacity"`	//仓库容量
	NeedUpdate		bool		`json:"-" xorm:"-"`
}

func (this *RoleRes) TableName() string {
	return "role_res"
}
