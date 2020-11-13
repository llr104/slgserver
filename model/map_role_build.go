package model

type MapRoleBuild struct {
	Id			int			`json:"id" xorm:"id pk autoincr"`
	RId			int			`json:"rid" xorm:"rid"`
	Type		int8		`json:"type"`
	Level		int			`json:"level"`
	X			int			`json:"x"`
	Y			int			`json:"y"`
	Name		string		`json:"name"`
	Wood		int			`json:"Wood"`
	Iron		int			`json:"iron"`
	Stone		int			`json:"stone"`
	Grain		int			`json:"grain"`
	Durable		int			`json:"durable"`
	Defender	int			`json:"defender"`
}

func (this *MapRoleBuild) TableName() string {
	return "map_role_build"
}
