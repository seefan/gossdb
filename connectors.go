package gossdb

import (
	"errors"

	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/ssdb_client"
)

//连接池
type Connectors struct {
	cfg    conf.Config //配置
	pool   []*Block
	index  int32
	length int32
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) Start(cfg *conf.Config) error {
	cfg.Default()
	if err := c.appendBlock(c.cfg.MaxPoolSize); err != nil {
		return err
	}
	return nil
}
func (c *Connectors) appendBlock(size int) error {
	var bs []*ssdb_client.SSDBClient
	for i := 0; i < size; i++ {
		if b, e := c.newSSDBClient(); e != nil {
			bs = append(bs, b)
		} else {
			//close bs
			for b = range bs {
				_ = b.Close()
			}
			return e
		}
	}
	c.pool = append(c.pool, newBlock(bs))
	c.length = int32(len(c.pool))
	return nil
}
func (c *Connectors) newSSDBClient() (*ssdb_client.SSDBClient, error) {
	b := &ssdb_client.SSDBClient{
		Host:             c.cfg.Host,
		Port:             c.cfg.Port,
		Password:         c.cfg.Password,
		WriteBufferSize:  c.cfg.WriteBufferSize,
		ReadBufferSize:   c.cfg.ReadBufferSize,
		ReadWriteTimeout: c.cfg.ReadWriteTimeout,
		WriteTimeout:     c.cfg.ReadWriteTimeout,
		ReadTimeout:      c.cfg.ReadWriteTimeout,
		ConnectTimeout:   c.cfg.ConnectTimeout,
		RetryEnabled:     c.cfg.RetryEnabled,
	}
	return b, b.Start()
}

//关闭连接池
func (c *Connectors) Close() {
	if c.pool != nil {
		//c.pool.Close()
	}
}

//状态信息
//
//  返回 string，一个详细连接池基本情况的字符串
func (c *Connectors) Info() string {
	if c.pool == nil {
		return "Please call the init function first"
	}
	return ""
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

//重置连接，重启连接，如果不成功，则将整个块设置为不可用
//
// element *PooledClient 要回收的连接
func (s *Block) resetClient(element *Client) {
	err := element.db.Close()
	if err != nil {
		s.lastError = err
	}
	err = element.db.Start()
	if err != nil {
		s.lastError = err
		s.available = false
	} else {
		s.setPoolClient(element)
	}
}

//关闭连接池
func (s *Block) Close() {
	for _, c := range s.pooled {
		if c != nil && c.db.IsOpen() {
			_ = c.db.Close()
		}
	}
	s.pooled = s.pooled[:0]
}
