package gossdb

var (
	//是否启动编码，启用后会对struct 进行 json 编码，以支持更多类型
	Encoding bool
	//配置密码, 之后将用于向服务器校验
	AuthPassword string
)

//根据配置初始化连接池
//
//  conf 连接池的初始化配置
//  password 配置密码，可选，用于向服务器校验
//  返回 一个可用的连接池
//  返回 err，可能的错误，操作成功返回 nil
//
//默认值
//
//	GetClientTimeout int 获取连接超时时间，单位为秒，默认1分钟
//	MaxPoolSize int 最大连接池个数，默认为10
//	MinPoolSize int 最小连接池数，默认为1
//	AcquireIncrement int  当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 3
//	MaxIdleTime int 最大空闲时间，指定秒内未使用则连接被丢弃。若为0则永不丢弃。默认值: 0
//  MaxWaitSize int 最大等待数目，当连接池满后，新建连接将排除等待池中连接释放，本值限制最大等待的数量。默认值: 1000
func NewPool(conf *Config, password ...string) (*Connectors, error) {
	if len(password) > 0 {
		AuthPassword = password[0]
	}
	//默认值处理
	c := new(Connectors)
	c.Init(conf)
	if err := c.Start(); err != nil {
		return nil, err
	}
	return c, nil
}
