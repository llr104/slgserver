package conn

type PushSync interface {
	IsCellView() bool		//是否格子视野
	BelongToRId() []int		//属于的rid
	PushMsgName() string	//推送名字
	Position() (int, int)	//x, y
	ToProto() interface{}	//转成proto
	Push() 					//推送
}
