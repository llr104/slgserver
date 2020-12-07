package proto

const (
	UnionChairman   		= 0	//盟主
	UnionViceChairman		= 1 //副盟主
	UnionCommon				= 2 //普通成员
)

const (
	UnionUntreated	= 1 //未处理
	UnionRefuse		= 1 //拒绝
	UnionAdopt   	= 2	//通过
	UnionHas   		= 3	//已经有联盟了
)

type Member struct {
	RId   int    `json:"rid"`
	Name  string `json:"name"`
	Title int8   `json:"title"`
}

type Union struct {
	Name   string   `json:"name"`   //联盟名字
	Cnt    int      `json:"cnt"`    //联盟人数
	Notice string   `json:"notice"` //公告
	Major  []Member `json:"major"` //联盟主要人物，盟主副盟主
}

type ApplyItem struct {
	Id       int    `json:"id"`
	RId      int    `json:"rid"`
	NickName string `json:"nick_name"`
}


//创建联盟
type CreateReq struct {
	Name	string	`json:"name"`
}

type CreateRsp struct {
	Id		int		`json:"id"`
	Name	string	`json:"name"`
}

//联盟列表
type ListReq struct {
}

type ListRsp struct {
	List	[]Union	`json:"list"`
}

//申请加入联盟
type JoinReq struct {
	Id	int		`json:"id"`
}

type JoinRsp struct {

}

//联盟成员
type MemberReq struct {
	Id	int		`json:"id"`
}

type MemberRsp struct {
	Id			int			`json:"id"`
	Members  	[]Member 	`json:"Members"`
}


//获取申请列表
type ApplyReq struct {
	Id	int		`json:"id"`
}

type ApplyRsp struct {
	Applys []ApplyItem `json:"applys"`
}

//审核
type VerifyReq struct {
	Id     int `json:"id"`		//申请操作的id
	Decide int `json:"decide"` 	//1是拒绝，2是通过
}

type VerifyRsp struct {
	Id     int `json:"id"`		//申请操作的id
	Decide int `json:"decide"` 	//1是拒绝，2是通过
}

//退出
type ExitReq struct {

}

type ExitRsp struct {

}


//解散
type DismissReq struct {

}

type DismissRsp struct {

}


type NoticeReq struct {
	Id	int		`json:"id"`
}

type NoticeRsp struct {
	Text 	string	`json:"text"`
}

//修改公告
type ModNoticeReq struct {
	Text 	string	`json:"text"`
}

type ModNoticeRsp struct {

}

//踢人
type KickReq struct {
	RId		int 	`json:"rid"`
}

type KickRsp struct {
	RId		int 	`json:"rid"`
}

//任命
type AppointReq struct {
	RId		int 	`json:"rid"`
	Title   int 	`json:"title"` //职位，1副盟主（目前只支持任命副盟主）
}

type AppointRsp struct {
	RId		int 	`json:"rid"`
	Title   int 	`json:"title"` //职位，1副盟主（目前只支持任命副盟主）
}
