package proto

type Facility struct {
	Name	string		`json:"name"`
	Level	int8		`json:"level"`
	Type	int8		`json:"type"`
}

type FacilitiesReq struct {
	CityId		int		`json:"cityId"`
}

type FacilitiesRsp struct {
	CityId		int			`json:"cityId"`
	Facilities	[]Facility	`json:"facilities"`
}

type UpFacilityReq struct {
	CityId int 	`json:"cityId"`
	FType  int8	`json:"fType"`
}

type UpFacilityRsp struct {
	CityId		int		`json:"cityId"`
	Facility	Facility`json:"facility"`
	RoleRes		RoleRes	`json:"role_res"`
}

type UpCityReq struct {
	CityId int 	`json:"cityId"`
}

type UpCityRsp struct {
	City	MapRoleCity	`json:"city"`
}