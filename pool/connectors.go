package pool

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/consts"
	"github.com/seefan/gossdb/ssdbclient"
)

//Connectors connection pool
//
//连接池
type Connectors struct {
	//当前连接池个数
	cellPos int32
	//状态 0：创建 1：正常 -1：关闭
	status byte
	//最大等待数量
	maxWait int32
	//连接池最大个数
	cellMax int32
	//连接池最小个数
	cellMin int32
	//pool
	cell []*Pool

	//This function is called when automatic serialization is performed, and it can be modified to use a custom serialization method
	//进行自动序列化时将调用这个函数，修改它可以使用自定义的序列化方式
	EncodingFunc func(v interface{}) []byte
	//config
	cfg *conf.Config
	//心跳检查
	//等待池
	poolWait chan *Client //连接池
	//最后的动作时间
	//
	//watchTicker
	watchTicker *time.Ticker
	//poolTemp
	clientTemp *sync.Pool
	//timer timp
	timerTemp *sync.Pool
	//round number
	round int32
	//性能指标
	//当前创建的连接数
	available int32
	//当前活跃连接数，运行中
	parallel int32
	//当前等待创建的连接数，等待
	waitCount int32
	//创建连接总耗时
	createSumTime int64
	//创建连接超时的个数
	createTimeout int32
	//创建连接的等待时间
	createWaitTime int64
	//当前返回失败的计数
	returnFail int32
	//运行总时长
	parallelSumTime int64
}

//NewConnectors initialize the connection pool using the configuration
//
//  @param cfg config
//
//使用配置初始化连接池
func NewConnectors(cfg *conf.Config) *Connectors {
	this := new(Connectors)
	this.cfg = cfg.Default()
	this.cellMax = int32(math.Floor(float64(cfg.MaxPoolSize) / float64(cfg.PoolSize)))
	this.cellMin = int32(math.Floor(float64(cfg.MinPoolSize) / float64(cfg.PoolSize)))
	this.maxWait = int32(cfg.MaxWaitSize)
	this.poolWait = make(chan *Client, cfg.MaxWaitSize)
	this.watchTicker = time.NewTicker(time.Second)
	this.cell = make([]*Pool, this.cellMax)

	this.EncodingFunc = func(v interface{}) []byte {
		if bs, err := json.Marshal(v); err == nil {
			return bs
		}
		return nil
	}
	this.clientTemp = &sync.Pool{
		New: func() interface{} {
			return &Client{Client: client.Client{}}
		},
	}
	this.timerTemp = &sync.Pool{
		New: func() interface{} {
			t := time.NewTimer(time.Duration(this.cfg.GetClientTimeout) * time.Second)
			t.Stop()
			return t
		},
	}
	this.status = consts.PoolStop
	return this
}

//后台的观察函数，处理连接池大小的扩展和收缩，连接池状态的检查等
//基本流程为先标记，再关闭
//标记的条件为如果活跃连接数不足，测试将连接池块长度缩减，然后检查该连接池块的连接有没有全部回收，如果全部回收就进行标记
//在下一个检查周期，将标记的块回收
//在检查周期过程中标记状态可能改变，如果块重用，将块内所有连接的状态检查一下，没有open的重新start一下
// func (c *Connectors) watchHealth() {
// 	for v := range c.watchTicker.C {
// 		atomic.StoreInt32(&c.available, 0)
// 		size := atomic.LoadInt32(&c.size)
// 		if c.minSize != c.maxSize && v.Second()%c.cfg.HealthSecond == 0 {
// 			parallel := atomic.LoadInt32(&c.parallel)
// 			if parallel < (size-1)*int32(c.cfg.PoolSize) && size-1 >= int32(c.minSize) {
// 				atomic.AddInt32(&c.size, -1)
// 			}
// 			c.watchPool(int(size))
// 		}
// 		waitCount := atomic.LoadInt32(&c.waitCount)
// 		if waitCount > 0 && size < int32(c.maxSize) {
// 			if err := c.appendPool(); err != nil {
// 				time.Sleep(time.Millisecond * 10)
// 			}
// 		}
// 		//println(c.Info())
// 	}
// }

//检查一下可关闭的连接池块，如果没有活动连接，可以关闭
// func (c *Connectors) watchPool(size int) {
// 	for i := size; i < c.maxSize; i++ {
// 		if c.pool[i] != nil {
// 			c.pool[i].CheckClose()
// 		}
// 	}
// 	for i := 0; i < size; i++ {
// 		if c.pool[i] != nil {
// 			c.pool[i].CheckClose()
// 		}
// 	}
// }

//初始化连接池
func (c *Connectors) appendPool() (err error) {
	pos := atomic.LoadInt32(&c.cellPos)
	if pos < c.cellMax {
		p := c.cell[pos]
		if p == nil {
			p = c.getPool()
			c.cell[pos] = p
		}
		if p.status != consts.PoolStart {
			if err = p.Start(); err != nil {
				return err
			}
		}
		p.index = pos
		atomic.AddInt32(&c.cellPos, 1)
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
				cc.close()
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
	c.cellPos = 0
	c.status = consts.PoolStart
	for i := c.cellPos; i < c.cellMin && err == nil; i++ {
		err = c.appendPool()
	}
	//go c.watchHealth()
	return
}

//回收Client
func (c *Connectors) closeClient(client *Client) {
	if c.status == consts.PoolStop {
		if client.SSDBClient.IsOpen() {
			_ = client.SSDBClient.Close()
		}
	} else {
		client.used = false
		ts := time.Now().UnixNano() - client.OpenTime
		atomic.AddInt64(&c.parallelSumTime, ts)
		pc := atomic.AddInt32(&c.parallel, -1)
		if client.SSDBClient.IsOpen() {
			waitCount := atomic.LoadInt32(&c.waitCount)
			if waitCount > 0 && pc%2 == 0 {
				c.poolWait <- client
			} else {
				client.pool.Set(client)
			}
		} else {
			client.pool.Set(client)
			atomic.StoreInt32(&client.pool.status, consts.PoolCheck)
			atomic.AddInt32(&c.returnFail, 1)
		}
	}
}

//GetClient gets an error-free connection and, if there is an error, returns when the connected function is called
//
// @return *Client
//
//获取一个无错误的连接，如果有错误，将在调用连接的函数时返回
func (c *Connectors) GetClient() *Client {
	cc, err := c.NewClient()
	//println("client get ", c.Info())
	if err == nil {
		return cc
	}
	cc = c.clientTemp.Get().(*Client)
	cc.Error = err
	return cc
}
func (c *Connectors) createClient() (cli *Client, err error) {
	//首先按位置，直接取连接，给n次机会
	size := atomic.LoadInt32(&c.cellPos)
	pi := atomic.LoadInt32(&c.round)

	for i := 0; i < 2; i++ {
		pi += int32(i)
		if pi >= size {
			pi = 0
		}
		p := c.cell[pi]
		if p.status != consts.PoolStop {
			cli = p.Get()
			if cli != nil {
				cli.Error = nil
				if p.health == consts.PoolCheck {
					if !cli.Ping() {
						err = cli.SSDBClient.Start()
					}
				} else if !cli.SSDBClient.IsOpen() {
					err = cli.SSDBClient.Start()
				}
				if err == nil {
					cli.used = true
					if c.round != cli.pool.index {
						atomic.StoreInt32(&c.round, cli.pool.index)
					}
					return cli, nil
				}
				p.Set(cli) //如果没有成功返回，就放回到连接池内
			}
		}
		runtime.Gosched()
	}
	return
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

	atomic.AddInt32(&c.available, 1)
	startTime := time.Now().UnixNano()
	cli, err = c.createClient()
	if cli != nil && err == nil {
		atomic.AddInt32(&c.parallel, 1)
		ts := time.Now().UnixNano() - startTime
		atomic.AddInt64(&c.createSumTime, ts)
		cli.OpenTime = startTime - ts
		return
	}

	//enter slow pool
	waitCount := atomic.LoadInt32(&c.waitCount)
	if waitCount >= c.maxWait {
		return nil, fmt.Errorf("pool is busy,Wait for connection creation has reached %d", waitCount)
	}
	waitCount = atomic.AddInt32(&c.waitCount, 1)
	timeout := c.timerTemp.Get().(*time.Timer)
	timeout.Reset(time.Duration(c.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout.C:
		atomic.AddInt32(&c.createTimeout, 1)
		err = fmt.Errorf("pool is busy,can not get new client in %d seconds,wait count is %d", c.cfg.GetClientTimeout, waitCount)
	case cli = <-c.poolWait:
		if cli == nil {
			err = errors.New("pool is Closed, can not get new client")
		} else {
			cli.used = true
			err = nil
			cli.OpenTime = time.Now().UnixNano()
			ts := cli.OpenTime - startTime
			atomic.AddInt64(&c.createWaitTime, ts) //等待时长
		}
	}
	atomic.AddInt32(&c.waitCount, -1)
	timeout.Stop()
	c.timerTemp.Put(timeout)
	return
}

//Close close connectors
//
//关闭连接池
func (c *Connectors) Close() {
	c.status = consts.PoolStop
	c.watchTicker.Stop()
	for _, cc := range c.cell {
		if cc != nil {
			cc.Close()
		}
	}
}

//Info returns connection pool status information
//
//  @return string
//
//返回连接池状态信息
func (c *Connectors) Info() string {
	available := atomic.LoadInt32(&c.available)
	createTime := atomic.LoadInt64(&c.createSumTime)
	parallelTime := atomic.LoadInt64(&c.parallelSumTime)
	inf := map[string]interface{}{
		"available": atomic.LoadInt32(&c.available),
		"parallel":  atomic.LoadInt32(&c.parallel),
		"waitCount": atomic.LoadInt32(&c.waitCount),
		//创建连接总耗时
		"createTime": atomic.LoadInt64(&c.createSumTime),
		//创建连接超时的个数
		"createTimeout": atomic.LoadInt32(&c.createTimeout),
		//创建连接的等待时间
		"createWaitTime": atomic.LoadInt64(&c.createWaitTime),
		//当前返回失败的计数
		"returnFail": atomic.LoadInt32(&c.returnFail),
		//运行总时长
		"parallelTime":    atomic.LoadInt64(&c.parallelSumTime),
		"parallelAvgTime": parallelTime / int64(available),
		"createAvgTime":   createTime / int64(available),
	}
	if bs, err := json.Marshal(inf); err == nil {
		return string(bs)
	}
	return "none"
}
