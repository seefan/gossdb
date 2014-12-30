package gossdb

//ssdb的连接配置
type Config struct {

	//ssdb的ip或主机名
	Host string
	// ssdb的端口
	Port int
	//获取连接超时时间，单位为秒，默认1分钟
	GetClientTimeout int
	//最大连接池个数，默认为10
	MaxPoolSize int
	//最小连接池数，默认为1
	MinPoolSize int
	//当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 3
	AcquireIncrement int
	//最大空闲时间，指定秒内未使用则连接被丢弃。若为0则永不丢弃。默认值: 0
	MaxIdleTime int
}
