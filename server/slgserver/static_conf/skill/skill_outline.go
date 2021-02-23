package skill

type trigger struct {
	Type int    `json:"type"`
	Des  string `json:"des"`
}

type triggerType struct {
	des string `json:"des"`
	list [] trigger `json:"list"`
}

type effect struct {
	Type   int    `json:"type"`
	Des    string `json:"des"`
	IsRate bool   `json:"isRate"`
}

type effectType struct {
	des string `json:"des"`
	list [] effect `json:"list"`
}

type target struct {
	Type   int    `json:"type"`
	Des    string `json:"des"`
}

type targetType struct {
	des string `json:"des"`
	list [] effect `json:"list"`
}


type outline struct {
	TriggerType triggerType `json:"trigger_type"`
	EffectType  effectType  `json:"effect_type"`
	TargetType  targetType  `json:"target_type"`
}
