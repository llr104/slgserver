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

type CreateRoleReq struct {
	UId			int		`json:"uid"`
	NickName 	string	`json:"nickName"`
	Sex			int8	`json:"sex"`
	SId			int		`json:"sid"`
	HeadId		int16	`json:"headId"`
}

type CreateRoleRsp struct {
	Role
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
}

type MapRoleCity struct {
	CityId		int			`json:"cityId"`
	RId			int			`json:"rid"`
	Name		string		`json:"name"`
	X			int			`json:"x"`
	Y			int			`json:"y"`
	IsMain		bool		`json:"is_main"`
	Level		int8		`json:"level"`
	Durable		int			`json:"durable"`
}

type MyCityReq struct {

}

type MyCityRsp struct {
	Citys []MapRoleCity `json:"citys"`
}
