package proto

type Conf struct {
	Type		int8		`json:"type"`
	Level		int			`json:"level"`
	Name		string		`json:"name"`
	Wood		int			`json:"Wood"`
	Iron		int			`json:"iron"`
	Stone		int			`json:"stone"`
	Grain		int			`json:"grain"`
	Durable		int			`json:"durable"`	//耐久
	Defender	int			`json:"defender"`	//防御等级
}

type ConfigReq struct {

}

type ConfigRsp struct {
	Confs []Conf
}

/*
type MapBaseBuild struct {
	Id		int		`json:"id"`
	X		int		`json:"x"`
	Y		int		`json:"y"`
	Type	int8	`json:"type"`
	Level	int		`json:"level"`
	Durable	int		`json:"durable"`
}
 */

type MapRoleBuild struct {
	Id			int			`json:"id"`
	RId			int			`json:"rid"`
	RNick		string 		`json:"RNick"` //角色昵称
	Name		string		`json:"name"`
	X			int    		`json:"x"`
	Y			int    		`json:"y"`
	Type		int8		`json:"type"`
	Level		int			`json:"level"`
	Durable		int			`json:"durable"`
	Defender	int			`json:"defender"`
}


type ScanReq struct {
	X 		int    		`json:"x"`
	Y 		int    		`json:"y"`
}

type ScanRsp struct {
	MRBuilds []MapRoleBuild `json:"mr_builds"` //角色建筑，包含被占领的基础建筑
	MCBuilds []RoleCity     `json:"mc_builds"` //角色城市
}

