package model

type MapRoleBuild struct {
	Id			int			`json:"id" xorm:"id pk autoincr"`
	RId			int			`json:"rid" xorm:"rid"`
	Type		int8		`json:"type"`
	Level		int8		`json:"level"`
	X			int			`json:"x"`
	Y			int			`json:"y"`
	Name		string		`json:"name"`
	Wood		int			`json:"Wood"`
	Iron		int			`json:"iron"`
	Stone		int			`json:"stone"`
	Grain		int			`json:"grain"`
	CurDurable	int			`json:"cur_durable"`
	MaxDurable	int			`json:"max_durable"`
	Defender	int			`json:"defender"`
}

func (this *MapRoleBuild) TableName() string {
	return "map_role_build"
}
