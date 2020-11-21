package model

type CityFacility struct {
	Id        	int		`xorm:"id pk autoincr"`
	CityId     	int		`xorm:"cityId"`
	Facilities 	string	`xorm:"facilities"`
	NeedUpdate	bool	`xorm:"-"`
}

func (this *CityFacility) TableName() string {
	return "city_facility"
}
