package skill

type level struct {
	Probability int   `json:"probability"`  //发动概率
	EffectValue []int `json:"effect_value"` //效果值
	EffectRound []int `json:"effect_round"` //效果持续回合数
}

type Conf struct {
	Name          string  `json:"name"`
	Trigger       int     `json:"trigger"` 			//发起类型
	Target        int     `json:"target"`  			//目标类型
	Des           string  `json:"des"`
	Limit         int     `json:"limit"`          	//可以被武将装备上限
	Arms          []int   `json:"arms"`           	//可以装备的兵种
	IncludeEffect []int   `json:"include_effect"` 	//技能包括的效果
	Levels        []level `json:"levels"`
}
