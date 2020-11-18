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
	CityId    	int     `json:"cityId"`
}

type MyGeneralReq struct {

}

type MyGeneralRsp struct {
	Generals []General `json:"generals"`
}


type Army struct {
	Id               int  `json:"id"`
	CityId           int  `json:"cityId"`
	Order            int8 `json:"order"`              //第几队，1-5队
	FirstId          int  `json:"firstId"`            //前锋武将id
	SecondId         int  `json:"secondId"`           //中锋武将id
	ThirdId          int  `json:"thirdId"`            //大营武将
	FirstSoldierCnt  int  `json:"first_soldier_cnt"`  //前锋士兵数量
	SecondSoldierCnt int  `json:"second_soldier_cnt"` //中锋士兵数量
	ThirdSoldierCnt  int  `json:"third_soldier_cnt"`  //大营士兵数量
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

