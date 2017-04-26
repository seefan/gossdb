package gossdb

import (
	"fmt"
	"github.com/seefan/goerr"
	"github.com/seefan/gopool"
	"sync"
	"time"
)

const (
	//连接池状态：创建
	PoolInit = 0
	//连接池状态：运行
	PoolStart = 1
	//连接池状态：关闭
	PoolStop = -1
)

//连接池
type Connectors struct {
	pool      chan *Client //连接池
	poolMap   *gopool.Pool
	cfg       *Config      //配置
	lock      sync.RWMutex //全局锁
	WaitCount int          //当前等待创建的连接数
	Status    int          //状态 0：创建 1：正常 -1：关闭
}

//用配置文件进行初始化
//
//  cfg 配置文件
func (c *Connectors) Init(cfg *Config) {
	c.setConfig(cfg)
	c.pool = make(chan *Client, cfg.MaxWaitSize)
	c.poolMap = gopool.NewPool()
	c.poolMap.WatchTime = c.cfg.HealthSecond
	c.poolMap.MinPoolSize = c.cfg.MinPoolSize
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) Start() error {
	c.WaitCount = 0
	for i := 0; i < c.cfg.MinPoolSize; i++ {
		cc := c.newClient(c)
		if err := cc.Start(); err != nil {
			return goerr.NewError(err, "启动连接池出错")
		}
		cc.Element=*c.poolMap.Append(cc)
	}
	c.Status = PoolStart
	return nil
}

//设置配置文件，主要是设置默认值
//
//  cfg 配置文件
func (c *Connectors) setConfig(conf *Config) {
	//默认值处理
	if conf.MaxPoolSize < 1 {
		conf.MaxPoolSize = 20
	}
	if conf.MinPoolSize < 1 {
		conf.MinPoolSize = 5
	}
	if conf.GetClientTimeout < 1 {
		conf.GetClientTimeout = 5
	}
	if conf.AcquireIncrement < 1 {
		conf.AcquireIncrement = 5
	}
	if conf.MaxWaitSize < 1 {
		conf.MaxWaitSize = 1000
	}
	if conf.HealthSecond < 1 {
		conf.HealthSecond = 5
	}

	if conf.MinPoolSize > conf.MaxPoolSize {
		conf.MinPoolSize = conf.MaxPoolSize
	}
	c.cfg = conf
}

//归还连接到连接池
//
//  cc 连接
func (c *Connectors) closeClient(cc *Client) {
	if cc == nil {
		return
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.Status == PoolStart {
		if cc.db.IsOpen() {
			if c.WaitCount > 0 { //有等待的连接
				c.pool <- cc
			} else {
				c.poolMap.Set(&cc.Element)
			}
		} else {
			c.poolMap.CloseClient(&cc.Element)
		}
	} else {
		if cc.db.IsOpen() {
			cc.db.Close()
		}
	}
}

//在连接池取一个新连接
//
//  返回 client，一个新的连接
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) NewClient() (client *Client, err error) {
	switch c.Status {
	case PoolStop:
		return nil, goerr.New("the Connectors is Closed, can not get new client.")
	case PoolInit:
		return nil, goerr.New("the Connectors is not inited, can not get new client.")
	}
	element, err := c.poolMap.Get()
	if err == nil {
		client = element.Value.(*Client)
		return
	}
	//get slow conn
	client, err = c.slowNew()
	if client != nil || err != nil {
		return client, err
	}
	//enter slow pool
	timeout := time.After(time.Duration(c.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		c.lock.Lock()
		c.WaitCount -= 1
		c.lock.Unlock()
		return nil, goerr.New("ssdb pool is busy,can not get new client in %d seconds", c.cfg.GetClientTimeout, c.Info())
	case cc := <-c.pool:
		if cc == nil {
			return nil, goerr.New("the Connectors is Closed, can not get new client.")
		}
		c.lock.Lock()
		c.WaitCount -= 1
		c.lock.Unlock()
		return cc, nil
	}

}

//检查是否可以快速创建新连接
//
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) slowNew() (*Client, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.WaitCount += 1
	if c.WaitCount > c.cfg.MaxWaitSize {
		c.WaitCount -= 1
		return nil, goerr.New("ssdb pool is busy,Wait for connection creation has reached %d", c.WaitCount)
	}
	if c.poolMap.Length < c.cfg.MaxPoolSize { //如果没有连接了，检查是否可以自动增加
		for i := 0; i < c.cfg.AcquireIncrement && c.poolMap.Length < c.cfg.MaxPoolSize; i++ {
			cc := c.newClient(c)
			if err := cc.Start(); err != nil {
				return nil, goerr.NewError(err, "扩展连接池出错")
			}
			cc.Element=*c.poolMap.Append(cc)
		}
		if element, err := c.poolMap.Get(); err == nil {
			println("slow new")
			return element.Value.(*Client), nil
		}
	}
	return nil, nil //没有慢速连接可用
}

//关闭连接池
func (c *Connectors) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.Status == PoolStart {
		close(c.pool)
	}
	c.Status = PoolStop
	//关闭连接池中的连接
	c.poolMap.Close()
	for len(c.pool) > 0 {
		cc := <-c.pool
		cc.Close()
	}
}

//状态信息
//
//  返回 string，一个详细连接池基本情况的字符串
func (c *Connectors) Info() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return fmt.Sprintf(`pool size:%d	actived client:%d	wait create:%d	config max pool size:%d	config Increment:%d`,
		c.poolMap.Length, c.poolMap.Current, c.WaitCount, c.cfg.MaxPoolSize, c.cfg.AcquireIncrement)
}

//创建一个新的连接
func (c *Connectors) newClient(p *Connectors) *Client {
	return &Client{
		db: &SSDBClient{
			host:     c.cfg.Host,
			port:     c.cfg.Port,
			password: c.cfg.Password,
		},
		pool: p,
	}
}
