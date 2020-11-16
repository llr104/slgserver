package proto

type Role struct {
	RId			int		`json:"rid"`
	UId			int		`json:"uid"`
	NickName 	string	`json:"nickName"`
	Sex			int8	`json:"sex"`
	SId			int		`json:"sid"`
	Balance		int		`json:"balance"`
	HeadId		int16	`json:"headId"`
	Profile		string	`json:"profile"`
}

type RoleRes struct {
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
}

type CreateRoleReq struct {
	UId			int		`json:"uid"`
	NickName 	string	`json:"nickName"`
	Sex			int8	`json:"sex"`
	SId			int		`json:"sid"`
	HeadId		int16	`json:"headId"`
}

type CreateRoleRsp struct {
	Role	Role	`json:"role"`
}

type RoleListReq struct {
	UId			int		`json:"uid"`
}

type RoleListRsp struct {
	Roles		[]Role
}

type EnterServerReq struct {
	SId			int		`json:"sid"`
}

type EnterServerRsp struct {
	SId			int		`json:"sid"`
	Role		Role	`json:"role"`
	RoleRes		RoleRes	`json:"role_res"`
}

type MapRoleCity struct {
	CityId		int			`json:"cityId"`
	RId			int			`json:"rid"`
	Name		string		`json:"name"`
	X			int			`json:"x"`
	Y			int			`json:"y"`
	IsMain		bool		`json:"is_main"`
	Level		int8		`json:"level"`
	CurDurable	int			`json:"cur_durable"`
	MaxDurable	int			`json:"max_durable"`
}

type MyCityReq struct {

}

type MyCityRsp struct {
	Citys []MapRoleCity `json:"citys"`
}

type MyRoleResReq struct {

}

type MyRoleResRsp struct {
	RoleRes		RoleRes	`json:"role_res"`
}