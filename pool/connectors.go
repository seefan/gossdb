package pool

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/consts"
	"github.com/seefan/gossdb/ssdbclient"
)

//Connectors
//
//连接池
type Connectors struct {
	cfg *conf.Config
	//心跳检查
	//等待池
	poolWait chan *Client //连接池
	//最后的动作时间
	//
	watchTicker *time.Ticker
	//连接池个数
	size byte
	//连接池最大个数
	maxSize int
	//连接池最小个数
	minSize int
	//状态 0：创建 1：正常 -1：关闭
	status byte
	//pool
	pool []*Pool
	//pos
	pos byte
	//处理等待状态的连接数
	waitCount int32
	//最大等待数量
	maxWaitSize int32
	//active
	activeCount int32
	//This function is called when automatic serialization is performed, and it can be modified to use a custom serialization method
	//进行自动序列化时将调用这个函数，修改它可以使用自定义的序列化方式
	EncodingFunc func(v interface{}) []byte
}

//NewConnectors initialize the connection pool using the configuration
//
//  @param cfg config
//
//使用配置初始化连接池
func NewConnectors(cfg *conf.Config) *Connectors {
	this := new(Connectors)
	this.cfg = cfg.Default()
	maxSize := int(math.Floor(float64(cfg.MaxPoolSize) / float64(cfg.PoolSize)))
	if maxSize > 256 {
		this.maxSize = 256
	} else {
		this.maxSize = maxSize
	}
	minSize := int(math.Floor(float64(cfg.MinPoolSize) / float64(cfg.PoolSize)))
	if minSize > 256 {
		this.minSize = 256
	} else {
		this.minSize = minSize
	}
	this.maxWaitSize = int32(cfg.MaxWaitSize)
	this.poolWait = make(chan *Client, cfg.MaxWaitSize)
	this.watchTicker = time.NewTicker(time.Second)
	this.pool = make([]*Pool, this.maxSize)
	this.EncodingFunc = func(v interface{}) []byte {
		if bs, err := json.Marshal(v); err == nil {
			return bs
		}
		return nil
	}
	this.status = consts.PoolStop
	return this
}

//后台的观察函数，处理连接池大小的扩展和收缩，连接池状态的检查等
func (c *Connectors) watchHealth() {
	for v := range c.watchTicker.C {
		size := int(c.size)
		if v.Second()%c.cfg.HealthSecond == 0 {
			activeCount := int(atomic.LoadInt32(&c.activeCount))
			if activeCount < (size-1)*c.cfg.PoolSize && size-1 >= c.minSize {
				c.size--
			}
			for i := size; i < c.maxSize; i++ {
				if c.pool[i] != nil {
					if c.pool[i].available.pos == c.pool[i].available.size && c.pool[i].status != consts.PoolStop {
						c.pool[i].status = consts.PoolStop
					}
					if c.pool[i].status == consts.PoolStop {
						c.pool[i].Close()
					}
				}
			}

			for i := 0; i < size; i++ {
				if c.pool[i].status == consts.PoolCheck {
					count := 0
					for _, c := range c.pool[i].pooled {
						if c.IsOpen() {
							count++
						}
					}
					if count == c.pool[i].size {
						c.pool[i].status = consts.PoolStart
					}
				}
			}
		}
		waitCount := atomic.LoadInt32(&c.waitCount)
		if waitCount > 0 && size < c.maxSize {
			if err := c.appendPool(); err != nil {
				time.Sleep(time.Millisecond * 10)
				_ = c.appendPool() //retry
			}
		}
	}
}

//初始化连接池
func (c *Connectors) appendPool() (err error) {
	size := int(c.size)
	if size < c.maxSize {
		p := c.pool[size]
		if p == nil {
			p = c.getPool()
			c.pool[size] = p
		}
		if p.status == consts.PoolStop {
			if err = p.Start(); err != nil {
				return err
			}
		}
		p.index = byte(size)
		c.size++
	}
	return nil
}

//获取一个连接池，关键点是设置关闭函数，用于处理自动回收
func (c *Connectors) getPool() *Pool {
	p := newPool(c.cfg.PoolSize)
	p.New = func() (*Client, error) {
		sc := ssdbclient.NewSSDBClient(c.cfg)
		err := sc.Start()
		if err != nil {
			return nil, err
		}
		sc.EncodingFunc = c.EncodingFunc
		cc := &Client{
			over: c,
			pool: p,
		}
		cc.Client = *client.NewClient(sc, c.cfg.AutoClose, func() {
			if cc.AutoClose {
				cc.Close()
			}
		})
		return cc, nil
	}
	return p
}

//Start start connectors
//
//  @return error，possible error, operation successfully returned nil
//
//启动连接池
func (c *Connectors) Start() (err error) {
	c.size = 0
	c.pos = 0
	c.status = consts.PoolStart
	for i := 0; i < c.minSize && err == nil; i++ {
		err = c.appendPool()
	}
	go c.watchHealth()
	return
}

//回收Client
func (c *Connectors) closeClient(client *Client) {
	client.used = false
	atomic.AddInt32(&c.activeCount, -1)
	if c.status == consts.PoolStop {
		if client.SSDBClient.IsOpen() {
			_ = client.SSDBClient.Close()
		}
	} else {
		if client.SSDBClient.IsOpen() {
			waitCount := atomic.LoadInt32(&c.waitCount)
			if waitCount > 0 && client.index%3 == 0 {
				c.poolWait <- client
			} else {
				client.pool.Set(client)
				c.pos = client.pool.index
			}
		} else {
			client.pool.Set(client)
			client.pool.status = consts.PoolCheck
		}
	}
}

//GetClient gets an error-free connection and, if there is an error, returns when the connected function is called
//
// @return *Client
//
//获取一个无错误的连接，如果有错误，将在调用连接的函数时返回
func (c *Connectors) GetClient() *Client {
	if cc, err := c.NewClient(); err == nil {
		return cc
	} else {
		return &Client{Client: client.Client{Error: err}}
	}
}

//NewClient take a new connection in the connection pool and return an error if there is an error
//
//  @return client new client
//  @return error possible error, operation successfully returned nil
//
//在连接池取一个新连接，如果出错将返回一个错误
func (c *Connectors) NewClient() (cli *Client, err error) {
	if c.status != consts.PoolStart {
		return nil, errors.New("connectors not start")
	}
	//首先按位置，直接取连接，给2次机会
	for i := 0; i < 2; i++ {
		pos := c.pos
		if pos >= c.size {
			pos = 0
		}
		p := c.pool[pos]
		if p.status != consts.PoolStop {
			cli = p.Get()
			if cli != nil {
				if p.status == consts.PoolCheck {
					if !cli.Ping() {
						err = cli.SSDBClient.Start()
					}
				} else if !cli.SSDBClient.IsOpen() {
					err = cli.SSDBClient.Start()
				}
				if err == nil {
					atomic.AddInt32(&c.activeCount, 1)
					cli.used = true
					return cli, nil
				}
				p.Set(cli) //如果没有成功返回，就放回到连接池内
			}
		}
		c.pos++
	}
	//enter slow pool
	waitCount := atomic.LoadInt32(&c.waitCount)
	if waitCount >= c.maxWaitSize {
		return nil, fmt.Errorf("pool is busy,Wait for connection creation has reached %d", waitCount)
	}
	atomic.AddInt32(&c.waitCount, 1)
	timeout := time.NewTimer(time.Duration(c.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout.C:
		err = fmt.Errorf("pool is busy,can not get new client in %d seconds,wait count is %d", c.cfg.GetClientTimeout, c.waitCount)
	case cli = <-c.poolWait:
		if cli == nil {
			err = errors.New("pool is Closed, can not get new client")
		} else {
			cli.used = true
			atomic.AddInt32(&c.activeCount, 1)
			err = nil
		}
	}
	atomic.AddInt32(&c.waitCount, -1)
	timeout.Stop()
	return
}

//Close close connectors
//
//关闭连接池
func (c *Connectors) Close() {
	c.status = consts.PoolStop
	c.watchTicker.Stop()
	for _, cc := range c.pool {
		if cc != nil {
			cc.Close()
		}
	}
	c.pool = c.pool[:0]
}

//Info returns connection pool status information
//
//  @return string
//
//返回连接池状态信息
func (c *Connectors) Info() string {
	return fmt.Sprintf("available is %d,wait is %d,connection count is %d,status is %d", c.activeCount, c.waitCount, int(c.size)*c.cfg.PoolSize, c.status)
}
