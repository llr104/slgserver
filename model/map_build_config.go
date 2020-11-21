package model

type MapBuildConfig struct {
	Id			int			`xorm:"id pk autoincr"`
	Type		int8		`xorm:"type"`
	Level		int8		`xorm:"level"`
	Name		string		`xorm:"name"`
	Wood		int			`xorm:"Wood"`
	Iron		int			`xorm:"iron"`
	Stone		int			`xorm:"stone"`
	Grain		int			`xorm:"grain"`
	Durable		int			`xorm:"durable"`
	Defender	int			`xorm:"defender"`

}

func (this *MapBuildConfig) TableName() string {
	return "map_build_config"
}

