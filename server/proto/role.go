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

