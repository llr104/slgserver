package proto

type Facility struct {
	Name	string		`json:"name"`
	MLevel	int8		`json:"mLevel"`
	CLevel	int8		`json:"cLevel"`
	Type	int8		`json:"type"`
}

type FacilitiesReq struct {
	CityId		int		`json:"cityId"`
}

type FacilitiesRsp struct {
	CityId		int			`json:"cityId"`
	Facilities	[]Facility	`json:"facilities"`
}