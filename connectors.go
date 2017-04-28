package gossdb

import (
	"github.com/seefan/gopool"
//	"github.com/seefan/gossdb/balance"
	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
	"sync"
)

//连接池
type Connectors struct {
	pool   *gopool.Pool //连接池
	cfg    conf.Config  //配置
	client sync.Pool    //客户端池
}

//用配置文件进行初始化
//
//  cfg 配置文件
func (c *Connectors) Init(cfg conf.Config) {
	c.client.New = func() interface{} {
		return &Client{
			pool: c,
		}
	}
	c.pool = gopool.NewPool()
	c.pool.NewClient = func() gopool.IClient {
		return &client.SSDBClient{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Password: cfg.Password,
		}
	}
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
func (c *Connectors) closeClient(cc *Client) {
	if cc == nil {
		return
	}
	c.pool.Set(cc.db.(*client.SSDBClient).Client.(*gopool.PooledClient))
	cc.db = nil
	c.client.Put(cc)
}
func (c *Connectors) NewClient() (*Client, error) {
	pc, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	sc := pc.Client.(*client.SSDBClient)
	sc.Client=pc
	cc := c.client.Get().(*Client)
	cc.db = client.IClient(sc)
	return cc, nil
}
