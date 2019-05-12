package conf

//ssdb连接池的配置
type Config struct {
	//ssdb的ip或主机名
	Host string
	//ssdb的端口
	Port int
	//获取连接超时时间，单位为秒。默认值: 5
	GetClientTimeout int
	//连接读写超时时间，单位为秒。默认值: 60
	ReadWriteTimeout int
	//连接写超时时间，单位为秒，如果不设置与ReadWriteTimeout会保持一致。默认值: 0
	WriteTimeout int
	//连接读超时时间，单位为秒，如果不设置与ReadWriteTimeout会保持一致。默认值: 0
	ReadTimeout int
	//最大连接个数。默认值: 100，PoolSize的整数倍，不足的话自动补足。
	MaxPoolSize int
	//最小连接个数。默认值: 20，PoolSize的整数倍，不足的话自动补足。
	MinPoolSize int
	//连接池最小单元连接数。默认值: 20，连接池增长连接时，以此值为步进值，可根据机器性能调整。
	PoolSize int
	//最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
	MaxWaitSize int
	//连接池内缓存的连接状态检查时间隔，单位为秒。默认值: 5
	HealthSecond int
	//连接的密钥
	Password string
	//连接写缓冲，默认为8k，单位为kb
	WriteBufferSize int
	//连接读缓冲，默认为8k，单位为kb
	ReadBufferSize int
	//是否启用重试，设置为true时，如果请求失败会再重试一次。
	RetryEnabled bool
	//创建连接的超时时间，单位为秒。默认值: 5
	ConnectTimeout int
	//auto close
	AutoClose bool
	//Automatic serialization of unknown types
	Encoding bool
}

// 设置默认配置
func (c *Config) Default() *Config {
	//默认值处理
	c.MaxPoolSize = defaultValue(c.MaxPoolSize, 100)
	c.PoolSize = defaultValue(c.PoolSize, 20)
	c.GetClientTimeout = defaultValue(c.GetClientTimeout, 5)
	c.MaxWaitSize = defaultValue(c.MaxWaitSize, 1000)
	c.HealthSecond = defaultValue(c.HealthSecond, 5)
	c.WriteBufferSize = defaultValue(c.WriteBufferSize, 8)
	c.ReadBufferSize = defaultValue(c.ReadBufferSize, 8)
	c.ReadWriteTimeout = defaultValue(c.ReadWriteTimeout, 60)
	c.ConnectTimeout = defaultValue(c.ConnectTimeout, 5)
	if c.MinPoolSize < c.PoolSize {
		c.MinPoolSize = c.PoolSize
	}
	if c.MaxPoolSize < c.PoolSize {
		c.MaxPoolSize = c.PoolSize
	}
	if c.MaxPoolSize == c.MinPoolSize {
		c.MaxPoolSize = c.MinPoolSize
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = c.ReadWriteTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = c.ReadWriteTimeout
	}
	return c
}

// 获取默认值
//
//  param，int，参数值
//  defaultValue，int，默认返回
//  返回，int。如果参数值小于1就返回默认值，否则返回参数值。
func defaultValue(param, defaultValue int) int {
	if param < 1 {
		return defaultValue
	} else {
		return param
	}
}
