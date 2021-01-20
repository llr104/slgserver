package proto

type Role struct {
	RId			int		`json:"rid"`
	UId			int		`json:"uid"`
	NickName 	string	`json:"nickName"`
	Sex			int8	`json:"sex"`
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
	Role Role `json:"role"`
}

type RoleListReq struct {
	UId			int		`json:"uid"`
}

type RoleListRsp struct {
	Roles		[]Role
}

type EnterServerReq struct {
	Session		string	`json:"session"`
}

type EnterServerRsp struct {
	Role    Role    `json:"role"`
	RoleRes RoleRes `json:"role_res"`
	Time    int64   `json:"time"`
	Token   string  `json:"token"`
}

type MapRoleCity struct {
	CityId     	int    	`json:"cityId"`
	RId        	int    	`json:"rid"`
	Name       	string 	`json:"name"`
	UnionId    	int    	`json:"union_id"` 	//联盟id
	UnionName  	string 	`json:"union_name"`	//联盟名字
	ParentId   	int    	`json:"parent_id"`	//上级id
	X          	int    	`json:"x"`
	Y          	int    	`json:"y"`
	IsMain     	bool   	`json:"is_main"`
	Level      	int8   	`json:"level"`
	CurDurable 	int    	`json:"cur_durable"`
	MaxDurable 	int    	`json:"max_durable"`
	OccupyTime	int64 	`json:"occupy_time"`
}

type MyCityReq struct {

}

type MyCityRsp struct {
	Citys []MapRoleCity `json:"citys"`
}

type MyRoleResReq struct {

}

type MyRoleResRsp struct {
	RoleRes RoleRes `json:"role_res"`
}

type MyRoleBuildReq struct {

}

type MyRoleBuildRsp struct {
	MRBuilds []MapRoleBuild `json:"mr_builds"` //角色建筑，包含被占领的基础建筑
}

/*
建筑发生变化
*/
type RoleBuildStatePush struct {
	MRBuild MapRoleBuild `json:"mr_build"` //角色建筑，包含被占领的基础建筑
}

type MyRolePropertyReq struct {

}

type MyRolePropertyRsp struct {
	RoleRes  RoleRes        `json:"role_res"`
	MRBuilds []MapRoleBuild `json:"mr_builds"` //角色建筑，包含被占领的基础建筑
	Generals []General      `json:"generals"`
	Citys    []MapRoleCity  `json:"citys"`
	Armys    []Army         `json:"armys"`
}

type UpPositionReq struct {
	X	int	`json:"x"`
	Y	int	`json:"y"`
}

type UpPositionRsp struct {
	X	int	`json:"x"`
	Y	int	`json:"y"`
}

type PosTag struct {
	X	int	`json:"x"`
	Y	int	`json:"y"`
}

type PosTagListReq struct {

}

type PosTagListRsp struct {
	PosTags	[]PosTag	`json:"pos_tags"`
}

type PosTagReq struct {
	Type	int `json:"type"`	//1是标记，0是取消标记
	X		int	`json:"x"`
	Y		int	`json:"y"`
}

type PosTagRsp struct {
	Type	int `json:"type"`	//1是标记，0是取消标记
	X		int	`json:"x"`
	Y		int	`json:"y"`
}
