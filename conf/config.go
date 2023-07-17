// Package conf gossdb config
package conf

// Config gossdb config
//
// ssdb连接池的配置
type Config struct {
	//connection key. Default is empty.
	//连接的密钥，默认为空
	Password string
	//ssdb hostname or ip. Default: 127.0.0.1
	//ssdb的ip或主机名，默认为: 127.0.0.1
	Host string
	//ssdb port. Default: 8888
	//ssdb的端口，默认为: 8888
	Port int
	//gets the connection timeout in seconds. Default: 5
	//获取连接超时时间，单位为秒。默认值: 5
	GetClientTimeout int
	//read/write timeout in seconds. Default: 60
	//连接读写超时时间，单位为秒。默认值: 60
	ReadWriteTimeout int
	//the connection write timeout, in seconds, is the same as the ReadWriteTimeout if not set. Default: 0
	//连接写超时时间，单位为秒，如果不设置与ReadWriteTimeout会保持一致。默认值: 0
	WriteTimeout int
	//the connection read timeout, in seconds, is the same as the ReadWriteTimeout if not set. Default: 0
	//连接读超时时间，单位为秒，如果不设置与ReadWriteTimeout会保持一致。默认值: 0
	ReadTimeout int
	//maximum number of connections. Default value: 100, integer multiple of PoolSize, if not enough, it will be filled automatically.
	//最大连接个数。默认值: 100，PoolSize的整数倍，不足的话自动补足。
	MaxPoolSize int
	//minimum number of connections. Default value: 20, integer multiple of PoolSize.
	//最小连接个数。默认值: 20，PoolSize的整数倍，不足的话自动补足。
	MinPoolSize int
	//minimum number of connection cells in the connection pool. Default value: 20. When the connection pool grows, this value is the step value, which can be adjusted according to the machine performance.
	//连接池一次创建连接的数据。默认值: 20，连接池增长连接时，以此值为步进值，可根据机器性能调整。
	PoolSize int
	//maximum number of waits. When the connection pool is full, the new connection can continue only after the connection in the pool is released. Default: 1000
	//最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
	MaxWaitSize int
	//the connection status check interval for the cache in the connection pool is in seconds. Default: 30
	//连接池内缓存的连接状态检查时间隔，单位为秒。默认值: 60
	HealthSecond int
	//connection write buffer, default 8k, in kb
	//连接写缓冲，默认为8k，单位为kb
	WriteBufferSize int
	//connection read buffer, default 8k, in kb
	//连接读缓冲，默认为8k，单位为kb
	ReadBufferSize int
	//the timeout for creating a connection in seconds. Default: 5
	//创建连接的超时时间，单位为秒。默认值: 5
	ConnectTimeout int
	//If the connection is automatically recycled, it will be recycled immediately after the connection operation is started. Default: false
	//是否自动回收连接，开启后连接使用操作一次后立即回收。默认值: false
	AutoClose bool
	//Automatic serialization of unknown types
	//是否自动进行序列化
	Encoding bool
	//if retry is enabled, set to true and try again if the request fails.
	//是否启用重试，设置为true时，如果连接状态异常会重新连接一次。
	RetryEnabled bool
}

// Default Gets the default configuration parameters
//
//	@return *Config
//
// 设置默认配置
func (c *Config) Default() *Config {
	//默认值处理
	c.Host = defaultString(c.Host, "127.0.0.1")
	c.Port = defaultValue(c.Port, 8888)
	c.MaxPoolSize = defaultValue(c.MaxPoolSize, 100)
	c.PoolSize = defaultValue(c.PoolSize, 20)
	c.GetClientTimeout = defaultValue(c.GetClientTimeout, 5)
	c.MaxWaitSize = defaultValue(c.MaxWaitSize, 1000)
	c.HealthSecond = defaultValue(c.HealthSecond, 60)
	c.WriteBufferSize = defaultValue(c.WriteBufferSize, 8)
	c.ReadBufferSize = defaultValue(c.ReadBufferSize, 8)
	c.ReadWriteTimeout = defaultValue(c.ReadWriteTimeout, 60)
	c.ConnectTimeout = defaultValue(c.ConnectTimeout, 5)
	if c.MinPoolSize < c.PoolSize {
		c.MinPoolSize = c.PoolSize
	}
	if c.MaxPoolSize < c.MinPoolSize {
		c.MaxPoolSize = c.MinPoolSize
	}
	c.ReadTimeout = defaultValue(c.ReadTimeout, c.ReadWriteTimeout)
	c.WriteTimeout = defaultValue(c.WriteTimeout, c.ReadWriteTimeout)
	return c
}

// 获取默认值
//
//	param，int，参数值
//	defaultValue，int，默认返回
//	返回，int。如果参数值小于1就返回默认值，否则返回参数值。
func defaultValue(param, defaultValue int) int {
	if param < 1 {
		return defaultValue
	}
	return param
}
func defaultString(param, defaultValue string) string {
	if param == "" {
		return defaultValue
	}
	return param
}
