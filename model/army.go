package model

//军队
type Army struct {
	Id               	int  	`json:"id" xorm:"id pk autoincr"`
	RId              	int  	`json:"rid" xorm:"rid"`
	CityId           	int  	`json:"cityId" xorm:"cityId"`
	Order            	int8 	`json:"order"`
	FirstId          	int 	`json:"firstId" xorm:"firstId"`
	SecondId         	int  	`json:"secondId" xorm:"secondId"`
	ThirdId          	int  	`json:"thirdId" xorm:"thirdId"`
	FirstSoldierCnt  	int  	`json:"first_soldier_cnt" xorm:"first_soldier_cnt"`
	SecondSoldierCnt 	int  	`json:"second_soldier_cnt" xorm:"second_soldier_cnt"`
	ThirdSoldierCnt  	int  	`json:"third_soldier_cnt" xorm:"third_soldier_cnt"`
	NeedUpdate			bool	`json:"-" xorm:"-"`
}

func (this *Army) TableName() string {
	return "army"
}


