package consts

const (
	//None init
	None = iota
	//PoolStart pool start
	//连接池状态：运行
	PoolStart
	//PoolStop pool stoped
	//连接池状态：关闭
	PoolStop
	//PoolCheck to check
	//连接池状态：检查
	PoolCheck
	//PoolStopping pool stoping
	//池连池状态：停止中
	PoolStopping
)
