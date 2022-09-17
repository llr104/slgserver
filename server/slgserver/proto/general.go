package proto

import "github.com/llr104/slgserver/server/slgserver/static_conf"

type GSkill struct {
	Id    int `json:"id"`
	Lv    int `json:"lv"`
	CfgId int `json:"cfgId"`
}
type General struct {
	Id            int       `json:"id"`
	CfgId         int       `json:"cfgId"`
	PhysicalPower int       `json:"physical_power"`
	Order         int8      `json:"order"`
	Level         int8      `json:"level"`
	Exp           int       `json:"exp"`
	CityId        int       `json:"cityId"`
	CurArms       int       `json:"curArms"`
	HasPrPoint    int       `json:"hasPrPoint"`
	UsePrPoint    int       `json:"usePrPoint"`
	AttackDis     int       `json:"attack_distance"`
	ForceAdded    int       `json:"force_added"`
	StrategyAdded int       `json:"strategy_added"`
	DefenseAdded  int       `json:"defense_added"`
	SpeedAdded    int       `json:"speed_added"`
	DestroyAdded  int       `json:"destroy_added"`
	StarLv        int8      `json:"star_lv"`
	Star          int8      `json:"star"`
	ParentId      int       `json:"parentId"`
	Skills        []*GSkill `json:"skills"`
	State         int8      `json:"state"`
}

func (this *General) ToArray() []int {
	r := make([]int, 0)

	r = append(r, this.Id)
	r = append(r, this.CfgId)
	r = append(r, this.PhysicalPower)
	r = append(r, int(this.Order))
	r = append(r, int(this.Level))
	r = append(r, this.Exp)
	r = append(r, this.CityId)
	r = append(r, this.CurArms)
	r = append(r, this.HasPrPoint)
	r = append(r, this.UsePrPoint)
	r = append(r, this.AttackDis)
	r = append(r, this.ForceAdded)
	r = append(r, this.StrategyAdded)
	r = append(r, this.SpeedAdded)
	r = append(r, this.DefenseAdded)
	r = append(r, this.DestroyAdded)
	r = append(r, int(this.StarLv))
	r = append(r, int(this.Star))
	return r
}

type MyGeneralReq struct {
}

type MyGeneralRsp struct {
	Generals []General `json:"generals"`
}

type ArmyListReq struct {
	CityId int `json:"cityId"`
}

type ArmyListRsp struct {
	CityId int    `json:"cityId"`
	Armys  []Army `json:"armys"`
}

type ArmyOneReq struct {
	CityId int  `json:"cityId"`
	Order  int8 `json:"order"`
}

type ArmyOneRsp struct {
	Army Army `json:"army"`
}

type Army struct {
	Id       int                         `json:"id"`
	CityId   int                         `json:"cityId"`
	UnionId  int                         `json:"union_id"` //联盟id
	Order    int8                        `json:"order"`    //第几队，1-5队
	Generals [static_conf.ArmyGCnt]int   `json:"generals"`
	Soldiers [static_conf.ArmyGCnt]int   `json:"soldiers"`
	ConTimes [static_conf.ArmyGCnt]int64 `json:"con_times"`
	ConCnts  [static_conf.ArmyGCnt]int   `json:"con_cnts"`
	Cmd      int8                        `json:"cmd"`
	State    int8                        `json:"state"` //状态:0:running,1:stop
	FromX    int                         `json:"from_x"`
	FromY    int                         `json:"from_y"`
	ToX      int                         `json:"to_x"`
	ToY      int                         `json:"to_y"`
	Start    int64                       `json:"start"` //出征开始时间
	End      int64                       `json:"end"`   //出征结束时间
}

//配置武将
type DisposeReq struct {
	CityId    int  `json:"cityId"`    //城市id
	GeneralId int  `json:"generalId"` //将领id
	Order     int8 `json:"order"`     //第几队，1-5队
	Position  int  `json:"position"`  //位置，-1到2,-1是解除该武将上阵状态
}

type DisposeRsp struct {
	Army Army `json:"army"`
}

//征兵
type ConscriptReq struct {
	ArmyId int   `json:"armyId"` //队伍id
	Cnts   []int `json:"cnts"`   //征兵人数
}

type ConscriptRsp struct {
	Army    Army    `json:"army"`
	RoleRes RoleRes `json:"role_res"`
}

//派遣队伍
type AssignArmyReq struct {
	ArmyId int  `json:"armyId"` //队伍id
	Cmd    int8 `json:"cmd"`    //命令：0:空闲 1:攻击 2：驻军 3:返回
	X      int  `json:"x"`
	Y      int  `json:"y"`
}

type AssignArmyRsp struct {
	Army Army `json:"army"`
}

//抽卡
type DrawGeneralReq struct {
	DrawTimes int `json:"drawTimes"` //抽卡次数
}

type DrawGeneralRsp struct {
	Generals []General `json:"generals"`
}

//合成
type ComposeGeneralReq struct {
	CompId int   `json:"compId"`
	GIds   []int `json:"gIds"` //合成材料
}

type ComposeGeneralRsp struct {
	Generals []General `json:"generals"`
}

//加点
type AddPrGeneralReq struct {
	CompId      int `json:"compId"`
	ForceAdd    int `json:"forceAdd"`
	StrategyAdd int `json:"strategyAdd"`
	DefenseAdd  int `json:"defenseAdd"`
	SpeedAdd    int `json:"speedAdd"`
	DestroyAdd  int `json:"destroyAdd"`
}

type AddPrGeneralRsp struct {
	Generals General `json:"general"`
}

type ConvertReq struct {
	GIds []int `json:"gIds"`
}

type ConvertRsp struct {
	GIds    []int `json:"gIds"`
	Gold    int   `json:"gold"`
	AddGold int   `json:"add_gold"`
}

type UpDownSkillReq struct {
	GId   int `json:"gId"`   //武将id
	CfgId int `json:"cfgId"` //技能配置id
	Pos   int `json:"pos"`   //位置0-2
}

type UpDownSkillRsp struct {
	GId   int `json:"gId"`   //武将id
	CfgId int `json:"cfgId"` //技能配置id
	Pos   int `json:"pos"`   //位置0-2
}

type LvSkillReq struct {
	GId int `json:"gId"` //武将id
	Pos int `json:"pos"` //位置0-2
}

type LvSkillRsp struct {
	GId int `json:"gId"` //武将id
	Pos int `json:"pos"` //位置0-2
}
