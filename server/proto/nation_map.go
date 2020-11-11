package proto

type Conf struct {
	Type		int8		`json:"type"`
	Name		string		`json:"name"`
	Wood		int			`json:"Wood"`
	Iron		int			`json:"iron"`
	Stone		int			`json:"stone"`
	Grain		int			`json:"grain"`
	Durable		int			`json:"durable"`
	Defender	int			`json:"defender"`
}

type ConfigReq struct {

}

type ConfigRsp struct {
	Confs []Conf
}

type baseBuild struct {
	Id		int		`json:"id"`
	X		int    	`json:"x"`
	Y		int    	`json:"y"`
	Type	int8	`json:"type"`
}

type roleBuild struct {
	Id		int			`json:"id"`
	RId		int			`json:"rid"`
	RNick	string 		`json:"RNick"` //角色昵称
	Name	string		`json:"name"`
	X		int    		`json:"x"`
	Y		int    		`json:"y"`
	Type	int8		`json:"type"`
	Durable	int			`json:"durable"`
}


type ScanReq struct {
	X 		int    		`json:"x"`
	Y 		int    		`json:"y"`
}

type ScanRsp struct {
	BBuilds	[]baseBuild	`json:"b_builds"`	//基础建筑
	RBuilds	[]roleBuild	`json:"n_builds"`	//角色建筑，包含被占领的基础建筑
}

