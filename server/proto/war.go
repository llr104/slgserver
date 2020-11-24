package proto

type WarReport struct {
	Id             	int    		`json:"id"`
	AttackRid      	int    		`json:"attack_rid"`
	DefenseRid     	int    		`json:"defense_rid"`
	BegAttackArmy  	string 		`json:"beg_attack_army"`
	BegDefenseArmy 	string 		`json:"beg_defense_army"`
	EndAttackArmy  	string 		`json:"end_attack_army"`
	EndDefenseArmy 	string 		`json:"end_defense_army"`
	AttackIsWin    	bool   		`json:"attack_is_win"`
	AttackIsRead  	bool   		`json:"attack_is_read"`
	DefenseIsRead  	bool   		`json:"defense_is_read"`
	DestroyDurable 	int			`json:"destroy_durable"`
	Occupy 			int			`json:"occupy"`
	X              	int       	`json:"x"`
	Y              	int       	`json:"y"`
	Time           	int64 		`json:"time"`
}

//战报推送
type WarReportPush struct {
	List	[]WarReport	`json:"list"`
}

type WarReportReq struct {

}

type WarReportRsp struct {
	List	[]WarReport	`json:"list"`
}

type WarReadReq struct {
	Id		uint		`json:"id"`
}

type WarReadRsp struct {
	Id		uint		`json:"id"`
}