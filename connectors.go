package gossdb

import (
	"errors"
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
	if cfg.ConnectTimeout < 1 {
		cfg.ConnectTimeout = 5
	}
	c.pool.NewClient = func() gopool.IClient {
		return &SSDBClient{
			Host:             cfg.Host,
			Port:             cfg.Port,
			Password:         cfg.Password,
			ReadBufferSize:   cfg.ReadBufferSize,
			WriteBufferSize:  cfg.WriteBufferSize,
			ReadWriteTimeout: cfg.ReadWriteTimeout,
			ConnectTimeout:   cfg.ConnectTimeout,
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
	c.pool.IdleTime = cfg.IdleTime
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) Start() error {
	if c.pool == nil {
		return errors.New("Please call the init function first")
	}
	return c.pool.Start()
}

//关闭连接池
func (c *Connectors) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

//状态信息
//
//  返回 string，一个详细连接池基本情况的字符串
func (c *Connectors) Info() string {
	if c.pool == nil {
		return "Please call the init function first"
	}
	return c.pool.Info()
}

//创建一个新连接
//
//    返回 *Client 可用的连接
//    返回 error 可能的错误
func (c *Connectors) NewClient() (*Client, error) {
	if c.pool == nil {
		return nil, errors.New("Please call the init function first")
	}
	pc, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	cc := pc.Client.(*SSDBClient)
	cc.client.cached = pc
	cc.client.db = cc
	if !cc.isOpen {
		return nil, errors.New("get client error")
	}
	cc.client.isActive = true
	return cc.client, nil
}
func (c *Connectors) closeClient(cc *Client) {
	cc.isActive = false
	c.pool.Set(cc.cached)
}
