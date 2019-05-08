package gossdb

import (
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
	size int32
	//连接池最大个数
	maxSize int
	//连接池最小个数
	minSize int
	//状态 0：创建 1：正常 -1：关闭
	status int
	//pool
	pool []*Pool
	//pos
	pos int32
	//处理等待状态的连接数
	waitCount int32
	//最大等待数量
	maxWaitSize int32
	//active
	activeCount int32
}

//用配置文件进行初始化
//
//  cfg 配置文件
func NewConnectors(cfg *conf.Config) *Connectors {
	this := new(Connectors)
	this.cfg = cfg.Default()
	this.maxSize = int(math.Floor(float64(cfg.MaxPoolSize) / float64(cfg.PoolSize)))
	this.minSize = int(math.Floor(float64(cfg.MinPoolSize) / float64(cfg.PoolSize)))
	this.maxWaitSize = int32(cfg.MaxWaitSize)
	this.poolWait = make(chan *Client, cfg.MaxWaitSize)
	this.watchTicker = time.NewTicker(time.Second)
	this.pool = make([]*Pool, cfg.MaxPoolSize)
	go this.watchHealth()
	this.status = consts.PoolStop
	return this
}
func (c *Connectors) watchHealth() {
	for v := range c.watchTicker.C {
		size := int(atomic.LoadInt32(&c.size))
		if v.Second()%c.cfg.HealthSecond == 0 {
			activeCount := int(atomic.LoadInt32(&c.activeCount))
			if activeCount < (size-1)*c.cfg.PoolSize && size-1 >= c.minSize {
				c.changeCount(&c.size, -1)
			}
			for i := size; i < c.cfg.MaxPoolSize; i++ {
				if c.pool[i] != nil {
					if c.pool[i].available.pos == c.pool[i].available.size && c.pool[i].Status != consts.PoolStop {
						c.pool[i].Status = consts.PoolStop
					}
					if c.pool[i].Status == consts.PoolStop {
						c.pool[i].Close()
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
	size := atomic.LoadInt32(&c.size)
	if int(size) < c.cfg.MaxPoolSize {
		p := c.pool[size]
		if p == nil {
			p = c.getPool()
			c.pool[size] = p
		}
		if p.Status == consts.PoolStop {
			if err = p.Start(); err != nil {
				return err
			}
		}
		p.index = size
		atomic.StoreInt32(&c.size, size+1)
	}
	return nil
}
func (c *Connectors) getPool() *Pool {
	p := newPool(c.cfg.PoolSize)
	p.New = func() (cc *Client, e error) {
		sc := ssdbclient.SSDBClient{
			Host:             c.cfg.Host,
			Port:             c.cfg.Port,
			Password:         c.cfg.Password,
			ReadWriteTimeout: c.cfg.ReadWriteTimeout,
			ReadTimeout:      c.cfg.ReadTimeout,
			WriteTimeout:     c.cfg.WriteTimeout,
			ReadBufferSize:   c.cfg.ReadBufferSize,
			WriteBufferSize:  c.cfg.WriteBufferSize,
			RetryEnabled:     c.cfg.RetryEnabled,
			ConnectTimeout:   c.cfg.ConnectTimeout,
		}
		if e = sc.Start(); e == nil {
			cn := &Client{
				Client: client.Client{SSDBClient: sc},
				over:   c,
				pool:   p,
			}

			cn.CloseMethod = func() {
				if cn.AutoClose {
					cn.Close()
				}
			}

			return cn, nil
		}
		return
	}
	return p
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) Start() (err error) {
	c.size = 0
	c.pos = 0
	c.status = consts.PoolStart
	for i := 0; i < c.minSize && err == nil; i++ {
		err = c.appendPool()
	}
	return
}

//关闭Client
func (c *Connectors) closeClient(client *Client) {
	c.changeCount(&c.activeCount, -1)
	if c.status == consts.PoolStop {
		if client.IsOpen() {
			_ = client.SSDBClient.Close()
		}
	} else {
		if client.IsOpen() {
			waitCount := atomic.LoadInt32(&c.waitCount)
			if waitCount > 0 && client.index%3 == 0 {
				c.poolWait <- client
			} else {
				client.pool.Set(client)
				atomic.StoreInt32(&c.pos, int32(client.pool.index))
			}
		} else {
			client.pool.Set(client)
			client.pool.Status = consts.PoolCheck
		}
	}
}
func (c *Connectors) changeCount(i32 *int32, d int32) {
	for {
		v := atomic.LoadInt32(i32)
		if atomic.CompareAndSwapInt32(i32, v, v+d) {
			break
		}
	}
}
func (c *Connectors) GetClient() *Client {
	if cc, err := c.NewClient(); err == nil {
		return cc
	} else {
		return &Client{Client: client.Client{},
			over: c,
		}
	}
}

//在连接池取一个新连接
//
//  返回 client，一个新的连接
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) NewClient() (cli *Client, err error) {
	if c.status != consts.PoolStart {
		return nil, errors.New("connectors not start")
	}
	//首先按位置，直接取连接，给2次机会
	for i := 0; i < 2; i++ {
		pos := atomic.LoadInt32(&c.pos)
		if pos >= int32(c.size) {
			pos = 0
		}
		p := c.pool[pos]
		if p.Status != consts.PoolStop {
			cli = p.Get()
			if cli != nil {
				if p.Status == consts.PoolCheck {
					if !cli.Ping() {
						err = cli.SSDBClient.Start()
					}
				}
				if err == nil {
					c.changeCount(&c.activeCount, 1)
					cli.isActive = true
					return cli, nil
				}
			}
		}
		atomic.CompareAndSwapInt32(&c.pos, pos, pos+1)
	}
	//enter slow pool
	waitCount := atomic.LoadInt32(&c.waitCount)
	if waitCount >= c.maxWaitSize {
		return nil, fmt.Errorf("pool is busy,Wait for connection creation has reached %d", waitCount)
	}
	c.changeCount(&c.waitCount, 1)
	timeout := time.After(time.Duration(c.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		err = fmt.Errorf("pool is busy,can not get new client in %d seconds,wait count is %d", c.cfg.GetClientTimeout, c.waitCount)
	case cli = <-c.poolWait:
		if cli == nil {
			err = errors.New("pool is Closed, can not get new client")
		} else {
			c.changeCount(&c.activeCount, 1)
			cli.isActive = true
			err = nil
		}
	}
	c.changeCount(&c.waitCount, -1)
	return
}

//***关闭连接池, 只修改连接池状态，让连接自行关闭，以免连接运行过程中被关闭
//关闭连接池
func (c *Connectors) Close() {
	c.status = consts.PoolStop
	for _, cc := range c.pool {
		if cc != nil {
			cc.Close()
		}
	}
	c.pool = c.pool[:0]
}

func (c *Connectors) Info() string {
	return fmt.Sprintf("available is %d,wait is %d,connection count is %d,status is %d", c.activeCount, c.waitCount, int(c.size)*c.cfg.PoolSize, c.status)
}
