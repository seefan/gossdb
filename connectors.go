package gossdb

import (
	"github.com/seefan/goerr"
	"github.com/seefan/gopool"
	"github.com/seefan/gossdb/conf"
)

//连接池
type Connectors struct {
	pool *gopool.Pool //连接池
	cfg  conf.Config  //配置
}

//用配置文件进行初始化
//
//  cfg 配置文件
func (c *Connectors) Init(cfg *conf.Config) {
	c.pool = gopool.NewPool()
	if cfg.WriteBufferSize < 1 {
		cfg.WriteBufferSize = 8
	}
	if cfg.ReadBufferSize < 1 {
		cfg.ReadBufferSize = 8
	}
	if cfg.ReadWriteTimeout < 1 {
		cfg.ReadWriteTimeout = 60
	}
	c.pool.NewClient = func() gopool.IClient {
		return &SSDBClient{
			Host:             cfg.Host,
			Port:             cfg.Port,
			Password:         cfg.Password,
			ReadBufferSize:   cfg.ReadBufferSize,
			WriteBufferSize:  cfg.WriteBufferSize,
			ReadWriteTimeout: cfg.ReadWriteTimeout,
			client: &Client{
				pool: c,
			},
		}
	}

	c.pool.GetClientTimeout = cfg.GetClientTimeout
	c.pool.MaxPoolSize = cfg.MaxPoolSize
	c.pool.MinPoolSize = cfg.MinPoolSize
	c.pool.AcquireIncrement = cfg.AcquireIncrement
	c.pool.MaxWaitSize = cfg.MaxWaitSize
	c.pool.HealthSecond = cfg.HealthSecond
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) Start() error {
	return c.pool.Start()
}

//关闭连接池
func (c *Connectors) Close() {
	c.pool.Close()
}

//状态信息
//
//  返回 string，一个详细连接池基本情况的字符串
func (c *Connectors) Info() string {
	return c.pool.Info()
}
func (c *Connectors) NewClient() (*Client, error) {
	pc, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	cc := pc.Client.(*SSDBClient)
	cc.client.cached = pc
	cc.client.db = cc
	if !cc.isOpen {
		return nil, goerr.New("get client error")
	}
	return cc.client, nil
}
func (c *Connectors) closeClient(cc *Client) {
	c.pool.Set(cc.cached)
}
