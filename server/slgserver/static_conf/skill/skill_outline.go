package skill

type trigger struct {
	Type int    `json:"type"`
	Des  string `json:"des"`
}

type triggerType struct {
	Des  string     `json:"des"`
	List [] trigger `json:"list"`
}

type effect struct {
	Type   int    `json:"type"`
	Des    string `json:"des"`
	IsRate bool   `json:"isRate"`
}

type effectType struct {
	Des  string    `json:"des"`
	List [] effect `json:"list"`
}

type target struct {
	Type   int    `json:"type"`
	Des    string `json:"des"`
}

type targetType struct {
	Des  string    `json:"des"`
	List [] target `json:"list"`
}


type outline struct {
	TriggerType triggerType `json:"trigger_type"`	//触发类型
	EffectType  effectType  `json:"effect_type"`	//效果类型
	TargetType  targetType  `json:"target_type"`	//目标类型
}
