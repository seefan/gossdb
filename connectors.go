package gossdb

import (
	"github.com/seefan/gossdb/balance"
	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
	"sync"
)

//连接池
type Connectors struct {
	pool   *balance.BalancePool //连接池
	cfg    conf.Config          //配置
	client *sync.Pool           //客户端池
}

//用配置文件进行初始化
//
//  cfg 配置文件
func (c *Connectors) Init(cfgs ...conf.Config) {
	c.client = &sync.Pool{
		New: func() interface{} {
			return &Client{
				pool: c,
			}
		},
	}
	c.pool = balance.NewBalancePool(cfgs)
	c.pool.NewSSDBClient = func(host string, port int, password string) *client.SSDBClient {
		return &client.SSDBClient{
			Host:     host,
			Port:     port,
			Password: password,
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
func (c *Connectors) NewClient() (*Client, error) {
	cc := c.client.Get().(*Client)

	cc.db = balance.NewProxyClient(c.pool)
	return cc, nil
}
func (c *Connectors) closeClient(cc *Client) {
	c.client.Put(cc)
}
