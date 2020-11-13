package model

type MapBuildConfig struct {
	Id			int			`json:"id" xorm:"id pk autoincr"`
	Type		int8		`json:"type"`
	Name		string		`json:"name"`
	Wood		int			`json:"Wood"`
	Iron		int			`json:"iron"`
	Stone		int			`json:"stone"`
	Grain		int			`json:"grain"`
	Durable		int			`json:"durable"`
	Defender	int			`json:"defender"`
}

func (this *MapBuildConfig) TableName() string {
	return "map_build_config"
}

