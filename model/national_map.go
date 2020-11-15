package model

type NationalMap struct {
	Id			int		`json:"id" xorm:"id pk autoincr"`
	MId			int		`json:"mid" xorm:"mid unique"`
	X			int		`json:"x"`
	Y			int		`json:"y"`
	Type		int8	`json:"type"`
	Level		int		`json:"level"`
	CurDurable	int		`json:"cur_durable"`
}

func (this *NationalMap) TableName() string {
	return "national_map"
}


