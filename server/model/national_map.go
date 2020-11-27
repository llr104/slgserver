package model

type NationalMap struct {
	Id			int		`xorm:"id pk autoincr"`
	MId			int		`xorm:"mid"`
	X			int		`xorm:"x"`
	Y			int		`xorm:"y"`
	Type		int8	`xorm:"type"`
	Level		int8	`xorm:"level"`
}

func (this *NationalMap) TableName() string {
	return "national_map"
}


