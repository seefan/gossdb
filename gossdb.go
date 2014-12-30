package gossdb

import (
	"fmt"
	"github.com/ssdb/gossdb/ssdb"
	"sync"
	"time"
)

type Connectors struct {
	pool        chan *Client //连接池
	cfg         *Config      //配置
	lock        sync.Mutex   //锁
	Size        int          //连接池大小
	ActiveCount int          //活动连接数
	Status      int          //状态 0：创建 1：正常 -1：关闭
}

//初始化连接池
func NewPool(conf *Config) (*Connectors, error) {
	//默认值处理
	if conf.MaxPoolSize < 1 {
		conf.MaxPoolSize = 10
	}
	if conf.MinPoolSize < 1 {
		conf.MinPoolSize = 1
	}
	if conf.GetClientTimeout <= 0 {
		conf.GetClientTimeout = 60
	}
	if conf.AcquireIncrement < 1 {
		conf.AcquireIncrement = 3
	}
	if conf.MinPoolSize > conf.MaxPoolSize {
		conf.MinPoolSize = conf.MaxPoolSize
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
	c.Status = 1
	return c, nil
}

//关闭连接池
func (this *Connectors) Close() {
	this.Status = -1
	for len(this.pool) > 0 {
		c := <-this.pool
		c.Client.Close()
	}

	close(this.pool)
}

//创建一个连接
func (this *Connectors) NewClient() (*Client, error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	switch this.Status {
	case -1:
		return nil, fmt.Errorf("连接池已关闭，无法获取新连接。")
	case 0:
		return nil, fmt.Errorf("连接池未初始化，无法获取新连接。")
	}
	if this.Size == this.ActiveCount && this.Size < this.cfg.MaxPoolSize { //如果没有连接了，检查是否可以自动增加
		for i := 0; i < this.cfg.AcquireIncrement && this.Size < this.cfg.MaxPoolSize; i++ {
			if c, err := this.newClient(); err == nil {
				this.pool <- c
				this.Size = len(this.pool) + this.ActiveCount
			} else { //如果新建连接有错误，就放弃
				break
			}
		}
	}
	timeout := time.Tick(time.Duration(this.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		return nil, fmt.Errorf("ssdb太忙，无法取得连接")
	case c := <-this.pool:
		this.ActiveCount += 1
		return c, nil
	}
}

//关闭一个连接，如果连接池关闭，就销毁连接，否则就回收到连接池
func (this *Connectors) closeClient(client *Client) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.Status != 1 {
		this.ActiveCount -= 1
		client.Client.Close()
	} else {
		client.lastTime = time.Now()
		this.pool <- client
	}
}

//按要求创建连接
func (this *Connectors) appendClient(size int) error {
	this.lock.Lock()
	defer this.lock.Unlock()
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
	db, err := ssdb.Connect(this.cfg.Host, this.cfg.Port)
	if err != nil {
		return nil, err
	}
	c := new(Client)
	c.Client = *db
	c.pool = this
	return c, nil
}
