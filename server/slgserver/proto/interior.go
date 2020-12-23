package proto

type CollectionReq struct {
}

type CollectionRsp struct {
}



type TransformReq struct {
	From		[]int	`json:"from"`		//0 Wood 1 Iron 2 Stone 3 Grain
	To		    []int	`json:"to"`		    //0 Wood 1 Iron 2 Stone 3 Grain
}

type TransformRsp struct {
}