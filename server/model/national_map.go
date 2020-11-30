package model

type NationalMap struct {
	MId			int		`xorm:"mid"`
	X			int		`xorm:"x"`
	Y			int		`xorm:"y"`
	Type		int8	`xorm:"type"`
	Level		int8	`xorm:"level"`
}


