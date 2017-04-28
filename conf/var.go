package conf

var (
	//ssdb的ip或主机名
	Host = "127.0.0.1"
	// ssdb的端口
	Port = 8888
	//获取连接超时时间，单位为秒。默认值: 5
	GetClientTimeout = 5
	//最大连接池个数。默认值: 20
	MaxPoolSize = 100
	//最小连接池数。默认值: 5
	MinPoolSize = 5
	//当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
	AcquireIncrement = 5
	//最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
	MaxWaitSize = 1000
	//连接池检查时间间隔
	HealthSecond = 5
	//默认配置文件名
	ConfigName = "config.ini"
)
