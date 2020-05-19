package consts

const (
	//None init
	None = iota
	//PoolStart pool start
	//连接池状态：运行
	PoolStart
	//PoolCheck to check
	//连接池状态：检查
	PoolCheck
	//PoolStop pool stop
	//连接池状态：关闭
	PoolStop
)
