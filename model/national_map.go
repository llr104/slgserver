package model

type NationalMap struct {
	Id		int			`json:"id" xorm:"id pk autoincr"`
	X		int			`json:"x"`
	Y		int			`json:"y"`
	Type	int8		`json:"type"`
	Level	int			`json:"level"`
}

func (this *NationalMap) TableName() string {
	return "national_map"
}


