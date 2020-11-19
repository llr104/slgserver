package proto

type General struct {
	Id        	int     `json:"id"`
	Name      	string  `json:"name"`
	CfgId     	int		`json:"cfgId"`
	Force     	int     `json:"force"`
	Strategy  	int     `json:"strategy"`
	Defense   	int     `json:"defense"`
	Speed     	int     `json:"speed"`
	Destroy   	int     `json:"destroy"`
	Cost      	int     `json:"cost"`
	Order     	int8    `json:"order"`
	Level		int8    `json:"level"`
	Exp			int		`json:"exp"`
	CityId    	int     `json:"cityId"`
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
	Id               int   `json:"id"`
	CityId           int   `json:"cityId"`
	Order            int8  `json:"order"`              //第几队，1-5队
	FirstId          int   `json:"firstId"`            //前锋武将id
	SecondId         int   `json:"secondId"`           //中锋武将id
	ThirdId          int   `json:"thirdId"`            //大营武将
	FirstSoldierCnt  int   `json:"first_soldier_cnt"`  //前锋士兵数量
	SecondSoldierCnt int   `json:"second_soldier_cnt"` //中锋士兵数量
	ThirdSoldierCnt  int   `json:"third_soldier_cnt"`  //大营士兵数量
	State            int8  `json:"state"`              //状态，0:空闲 1:攻击 2：驻军 3:返回
	FromX            int   `json:"from_x"`
	FromY            int   `json:"from_y"`
	ToX              int   `json:"to_x"`
	ToY              int   `json:"to_y"`
	Start            int64 `json:"start"`//出征开始时间
	End              int64 `json:"end"`//出征结束时间
}

//配置武将
type DisposeReq struct {
	CityId		int     `json:"cityId"`		//城市id
	GeneralId	int     `json:"generalId"`	//将领id
	Order		int8	`json:"order"`		//第几队，1-5队
	Position	int8	`json:"position"`	//位置，0-3,0是解除该武将上阵状态
}

type DisposeRsp struct {
	Army	Army	`json:"army"`
}

//征兵
type ConscriptReq struct {
	ArmyId		int  	`json:"armyId"`		//队伍id
	FirstCnt	int		`json:"firstCnt"`	//前锋征兵人数
	SecondCnt	int		`json:"secondCnt"`	//中锋征兵人数
	ThirdCnt	int		`json:"thirdCnt"`	//大营征兵人数
}

type ConscriptRsp struct {
	Army	Army	`json:"army"`
	RoleRes	RoleRes	`json:"role_res"`
}

//派遣队伍
type AssignArmyReq struct {
	ArmyId int  `json:"armyId"` //队伍id
	State  int8 `json:"state"`  //状态，0:空闲 1:攻击 2：驻军 3:返回
	X      int  `json:"x"`
	Y      int  `json:"y"`
}

type AssignArmyRsp struct {
	Army		Army	`json:"army"`
}
