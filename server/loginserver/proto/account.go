package proto

type LoginReq struct {
	Username 	string    	`json:"username"`
	Password 	string    	`json:"password"`
	Ip		 	string		`json:"ip"`
	Hardware	string		`json:"hardware"`
}

type LoginRsp struct {
	Username 	string    	`json:"username"`
	Password 	string    	`json:"password"`
	Session	 	string		`json:"session"`
	UId			int			`json:"uid"`
}

type ReLoginReq struct {
	Session 	string	`json:"session"`
	Ip			string	`json:"ip"`
	Hardware	string	`json:"hardware"`
}

type ReLoginRsp struct {
	Session string    	`json:"session"`
}

type LogoutReq struct {
	UId      int		`json:"uid"`
}

type LogoutRsp struct {
	UId      int		`json:"uid"`
}

type Server struct {
	Id		int			`json:"id"`
	Slg		string		`json:"slg"`
	Chat	string		`json:"chat"`
}

type ServerListReq struct {

}

type ServerListRsp struct {
	Lists	[]Server
}
