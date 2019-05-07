package pool

import (
	"fmt"
	//	"fmt"
	"sync"
	"sync/atomic"
	"time"

	//"github.com/seefan/goerr"

	"github.com/golangteam/function/errors"
	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/ssdb_client"
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
	cfg *conf.Config
	//心跳检查
	//等待池
	poolWait chan *Client //连接池
	//最后的动作时间
	lastTime int64
	//连接池个数
	size int
	//状态 0：创建 1：正常 -1：关闭
	status int
	//pool
	pool []*Pool
	//pos
	pos int32
	//lock
	lock sync.Mutex
	//处理等待状态的连接数
	waitCount int32
	//最大等待数量
	maxWaitSize int32
}

//用配置文件进行初始化
//
//  cfg 配置文件
func NewConnectors(cfg *conf.Config) *Connectors {
	this := new(Connectors)
	this.cfg = cfg.Default()
	this.maxWaitSize = int32(cfg.MaxWaitSize)
	this.poolWait = make(chan *Client, cfg.MaxWaitSize)
	return this
}

//初始化连接池
func (c *Connectors) appendPool() (err error) {
	p := c.getPool()
	if err = p.Start(); err == nil {
		p.index = int32(len(c.pool))
		c.pool = append(c.pool, p)
		c.size = int(p.index + 1)
		return nil
	} else {
		return err
	}
}
func (c *Connectors) getPool() *Pool {
	p := newPool(c.cfg.PoolSize)
	p.New = func() (cc *Client, e error) {
		sc := ssdb_client.SSDBClient{
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
			cc := &Client{
				Client: client.Client{SSDBClient: sc},
				over:   c,
				pool:   p,
			}
			return cc, nil
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
	c.status = PoolStart
	for i := 0; i < c.cfg.PoolNumber; i++ {
		err = c.appendPool()
	}
	return
}

//关闭Client
func (c *Connectors) closeClient(client *Client) {
	atomic.StoreInt64(&c.lastTime, time.Now().Unix())
	if c.status == PoolStart {
		atomic.StoreInt32(&c.pos, int32(client.pool.index))
		waitCount := atomic.LoadInt32(&c.waitCount)
		if waitCount > 0 && client.index%3 == 0 {
			c.poolWait <- client
		} else {
			client.pool.Set(client)
		}
	} else {
		if client.IsOpen() {
			_ = client.SSDBClient.Close()
		}
	}
}

//在连接池取一个新连接
//
//  返回 client，一个新的连接
//  返回 err，可能的错误，操作成功返回 nil
func (c *Connectors) NewClient() (cc *Client, err error) {
	if c.status != PoolStart {
		return nil, errors.New("connectors not start")
	}
	//首先按位置，直接取连接，给3次机会
	for i := 0; i < 2; i++ {
		pos := atomic.LoadInt32(&c.pos)
		if pos >= int32(c.size) {
			pos = 0
		}
		cc, err = c.pool[pos].Get()
		if err == nil {
			return
		}
		atomic.CompareAndSwapInt32(&c.pos, pos, pos+1)
	}
	//enter slow pool
	waitCount := atomic.LoadInt32(&c.waitCount)
	if waitCount >= c.maxWaitSize {
		return nil, fmt.Errorf("pool is busy,Wait for connection creation has reached %d", waitCount)
	}
	for {
		v := atomic.LoadInt32(&c.waitCount)
		if atomic.CompareAndSwapInt32(&c.waitCount, v, v+1) {
			break
		}
	}
	timeout := time.After(time.Duration(c.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		err = fmt.Errorf("pool is busy,can not get new client in %d seconds,wait count is %d", c.cfg.GetClientTimeout, c.waitCount)
	case cw := <-c.poolWait:
		if cw == nil {
			err = errors.New("pool is Closed, can not get new client")
		} else {
			cc = cw
			err = nil
		}
	}
	for {
		v := atomic.LoadInt32(&c.waitCount)
		if atomic.CompareAndSwapInt32(&c.waitCount, v, v-1) {
			break
		}
	}
	return
}

//***关闭连接池, 只修改连接池状态，让连接自行关闭，以免连接运行过程中被关闭
//关闭连接池
func (c *Connectors) Close() {
	c.status = PoolStop
	for _, cc := range c.pool {
		cc.Close()
	}
	c.pool = c.pool[:0]
}

//func (c *Connectors) Info() string {
//	return fmt.Sprintf("available is %d,size is %d,status is %d", this.activeCount, this.size, this.status)
//}
