package conn

type PushSync interface {
	IsCellView() bool				//是否格子视野
	IsCanView(rid, x, y int)bool	//是否能出现在视野
	BelongToRId() []int				//属于的rid
	PushMsgName() string			//推送名字
	Position() (int, int)			//x, y
	TPosition() (int, int)			//目标x, y
	ToProto() interface{}			//转成proto
	Push() 							//推送
}
