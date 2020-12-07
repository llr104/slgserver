package proto

const (
	UnionChairman   		= 0	//盟主
	UnionViceChairman		= 1 //副盟主
	UnionCommon				= 2 //普通成员
)

type Member struct {
	Rid   int    `json:"rid"`
	Name  string `json:"name"`
	Title int8   `json:"title"`
}

type Union struct {
	Name   string   `json:"name"`   //联盟名字
	Cnt    int      `json:"cnt"`    //联盟人数
	Notice string   `json:"notice"` //公告
	Major  []Member `json:"major"` //联盟主要人物，盟主副盟主
}

type CreateReq struct {
	Name	string	`json:"name"`
}

type CreateRsp struct {
	Id		int		`json:"id"`
	Name	string	`json:"name"`
}

type ListReq struct {
}

type ListRsp struct {
	List	[]Union	`json:"list"`
}

type JoinReq struct {
	Id	int		`json:"id"`
}

type JoinRsp struct {

}

type MemberReq struct {
	Id	int		`json:"id"`
}

type MemberRsp struct {
	Id			int			`json:"id"`
	Members  	[]Member 	`json:"Members"`
}