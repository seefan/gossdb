package gossdb

//ssdb连接池的配置
type Config struct {

	//ssdb的ip或主机名
	Host string
	// ssdb的端口
	Port int
	//获取连接超时时间，单位为秒。默认值: 5
	GetClientTimeout int
	//最大连接池个数。默认值: 20
	MaxPoolSize int
	//最小连接池数。默认值: 5
	MinPoolSize int
	//当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
	AcquireIncrement int
	//最大空闲时间，指定秒内未使用则连接被丢弃。若为0则永不丢弃。默认值: 0
	MaxIdleTime int
	//最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
	MaxWaitSize int
	//连接池内缓存的连接状态检查时间隔，单位为秒。默认值: 300
	HealthSecond int
	//数据库状态检查，单位为秒。默认值: 5
	DBHealthSecond int
}
