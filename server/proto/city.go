package proto

type FacilitiesReq struct {
	CityId		int		`json:"cityId"`
}

type FacilitiesRsp struct {
	CityId     int    `json:"cityId"`
	Facilities string `json:"facilities"`
}