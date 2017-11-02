package gopool

//连接接口
//
// 连接池内的连接结构只要实现这个接口就可以嵌入池内使用
type IClient interface {
	//打开连接
	//
	// 返回，error。如果连接到服务器时出错，就返回错误信息，否则返回nil
	Start() error
	//关闭连接
	//
	// 返回，error。如果关闭连接时出错，就返回错误信息，否则返回nil
	Close() error
	//是否打开
	//
	// 返回，bool。如果已连接到服务器，就返回true。
	IsOpen() bool
	//检查连接状态
	//
	// 返回，bool。如果无法访问服务器，就返回false。
	Ping() bool
}

//缓存的连接
//
//内部使用
type PooledClient struct {
	//pos
	index int
	// The poolWait to which this element belongs.
	pool *Pool
	//value
	Client IClient
	//used status
	isUsed bool
	//last time
	lastTime int64
}