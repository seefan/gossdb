package gossdb

import (
	"fmt"
	"github.com/ssdb/gossdb/ssdb"
	"sync"
	"time"
)

type Connectors struct {
	pool chan *Client //连接池
	cfg  *Config      //配置
	lock sync.RWMutex //锁
	Size int          //连接池大小
}

//初始化连接池
func NewPool(conf *Config) (*Connectors, error) {
	if conf.MaxPoolSize < 1 {
		conf.MaxPoolSize = 1
	}
	if conf.MinPoolSize > conf.MaxPoolSize {
		conf.MinPoolSize = conf.MaxPoolSize
	}
	if conf.GetClientTimeout == 0 {
		conf.GetClientTimeout = 10
	}
	c := &Connectors{
		pool: make(chan *Client, conf.MaxPoolSize),
		cfg:  conf,
	}
	c.appendClient(conf.MinPoolSize)
	c.Size = len(c.pool)
	if c.Size == 0 {
		return nil, fmt.Errorf("创建连接池失败，无法取得连接。")
	}
	return c, nil
}

//关闭连接池
func (this *Connectors) Close() {
	close(this.pool)
}

//创建一个连接
func (this *Connectors) NewClient() (*Client, error) {
	this.lock.RLock()
	this.lock.RUnlock()
	timeout := time.Tick(time.Duration(this.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		return nil, fmt.Errorf("ssdb太忙，无法取得连接")
	case c := <-this.pool:
		return c, nil
	}
}

//按要求创建连接
func (this *Connectors) appendClient(size int) error {
	for i := 0; i < size; i++ {
		if nc, err := this.newClient(); err == nil {
			this.pool <- nc
		} else {
			return err
		}
	}
	return nil
}

//创建一个连接
func (this *Connectors) newClient() (*Client, error) {
	this.lock.Lock()
	this.lock.Unlock()
	db, err := ssdb.Connect(this.cfg.Host, this.cfg.Port)
	if err != nil {
		return nil, err
	}
	c := new(Client)
	c.Client = *db
	c.pool = this
	return c, nil
}
