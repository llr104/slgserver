package proto


type Skill struct {
	Id             	int 	`json:"id"`
	CfgId          	int 	`json:"cfgId"`
	Generals 		[]int 	`json:"generals"`
}

type SkillListReq struct {

}

type SkillListRsp struct {
	List []Skill `json:"list"`
}