package proto

type General struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CfgId     int		`json:"cfgId"`
	Force     int       `json:"force"`
	Strategy  int       `json:"strategy"`
	Defense   int       `json:"defense"`
	Speed     int       `json:"speed"`
	Destroy   int       `json:"destroy"`
	Cost      int       `json:"cost"`
	ArmyId    int       `json:"armyId"`
	CityId    int       `json:"cityId"`
}

type MyGeneralReq struct {

}

type MyGeneralRsp struct {
	Generals []General `json:"generals"`
}
