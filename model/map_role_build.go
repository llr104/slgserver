package model

type MapRoleBuild struct {
	Id			int			`xorm:"id pk autoincr"`
	RId			int			`xorm:"rid"`
	Type		int8		`xorm:"type"`
	Level		int8		`xorm:"level"`
	X			int			`xorm:"x"`
	Y			int			`xorm:"y"`
	Name		string		`xorm:"name"`
	Wood		int			`xorm:"Wood"`
	Iron		int			`xorm:"iron"`
	Stone		int			`xorm:"stone"`
	Grain		int			`xorm:"grain"`
	CurDurable	int			`xorm:"cur_durable"`
	MaxDurable	int			`xorm:"max_durable"`
	Defender	int			`xorm:"defender"`
	NeedUpdate	bool		`xorm:"-"`
}

func (this *MapRoleBuild) TableName() string {
	return "map_role_build"
}
