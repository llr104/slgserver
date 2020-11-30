package proto

type General struct {
	Id        		int     `json:"id"`
	CfgId     		int		`json:"cfgId"`
	PhysicalPower 	int     `json:"physical_power"`
	Cost      		int     `json:"cost"`
	Order     		int8    `json:"order"`
	Level			int8    `json:"level"`
	Exp				int		`json:"exp"`
	CityId    		int     `json:"cityId"`
}

type MyGeneralReq struct {

}

type MyGeneralRsp struct {
	Generals []General `json:"generals"`
}


type ArmyListReq struct {
	CityId	int  `json:"cityId"`
}

type ArmyListRsp struct {
	CityId	int  	`json:"cityId"`
	Armys	[]Army	`json:"armys"`
}


type Army struct {
	Id       	int   	`json:"id"`
	CityId   	int   	`json:"cityId"`
	Order    	int8  	`json:"order"`   //第几队，1-5队
	Generals	[]int 	`json:"generals"`
	Soldiers 	[]int 	`json:"soldiers"`
	Cmd      	int8  	`json:"cmd"` //命令，0:空闲 1:攻击 2：驻军 3:返回
	State		int8	`json:"state"` //状态:0:running,1:stop
	FromX    	int   	`json:"from_x"`
	FromY    	int   	`json:"from_y"`
	ToX      	int   	`json:"to_x"`
	ToY      	int   	`json:"to_y"`
	Start    	int64 	`json:"start"`//出征开始时间
	End      	int64 	`json:"end"`//出征结束时间
}

//配置武将
type DisposeReq struct {
	CityId		int     `json:"cityId"`		//城市id
	GeneralId	int     `json:"generalId"`	//将领id
	Order		int8	`json:"order"`		//第几队，1-5队
	Position	int		`json:"position"`	//位置，-1到2,-1是解除该武将上阵状态
}

type DisposeRsp struct {
	Army	Army	`json:"army"`
}

//征兵
type ConscriptReq struct {
	ArmyId		int  	`json:"armyId"`		//队伍id
	Cnts		[]int	`json:"cnts"`		//征兵人数
}

type ConscriptRsp struct {
	Army	Army	`json:"army"`
	RoleRes	RoleRes	`json:"role_res"`
}

//派遣队伍
type AssignArmyReq struct {
	ArmyId int  `json:"armyId"` //队伍id
	Cmd    int8 `json:"cmd"`  //命令：0:空闲 1:攻击 2：驻军 3:返回
	X      int  `json:"x"`
	Y      int  `json:"y"`
}

type AssignArmyRsp struct {
	Army		Army	`json:"army"`
}


//抽卡
type DrawGeneralReq struct {
	DrawTimes int  `json:"drawTimes"` //抽卡次数
}

type DrawGeneralRsp struct {
	Generals []General `json:"generals"`
}