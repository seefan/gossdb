package consts

const (
	None = iota
	//pool is busy
	PoolEmpty
	//poos is not start
	PoolNotStart
	//连接池状态：关闭
	PoolStop
	//连接池状态：运行
	PoolStart
	//连接池状态：检查
	PoolCheck
)
