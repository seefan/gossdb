package gopool

import (
	"fmt"
	"sync"
	"time"

	"github.com/seefan/goerr"
)

const (
	//连接池状态：创建
	PoolInit = 0
	//连接池状态：运行
	PoolStart = 1
	//连接池状态：关闭
	PoolStop = -1
	//连接池状态：重启
	PoolReStart = 2
)

// 连接池结构
type Pool struct {
	//处理等待状态的连接数
	waitCount int
	//element list
	pooled *Slice
	//等待池
	poolWait chan *PooledClient //连接池

	//create new Closed
	NewClient func() IClient
	//状态
	Status int
	//config
	//获取连接超时时间，单位为秒。默认值: 5
	GetClientTimeout int
	//最大连接池个数。默认值: 20
	MaxPoolSize int
	//最小连接池数。默认值: 5
	MinPoolSize int
	//当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
	AcquireIncrement int
	//最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
	MaxWaitSize int
	//连接池内缓存的连接状态检查时间隔，单位为秒。默认值: 5
	HealthSecond int
	//连接空闲时间，超过这个时间可能会被回收，单位为秒。默认值:60
	IdleTime int
	//watch
	watcher *time.Ticker
	//lock
	lock sync.RWMutex
}

// 设置默认配置
func (p *Pool) defaultConfig() {
	//默认值处理
	p.MaxPoolSize = defaultValue(p.MaxPoolSize, 20)
	p.MinPoolSize = defaultValue(p.MinPoolSize, 5)
	p.GetClientTimeout = defaultValue(p.GetClientTimeout, 5)
	p.AcquireIncrement = defaultValue(p.AcquireIncrement, 5)
	p.MaxWaitSize = defaultValue(p.MaxWaitSize, 1000)
	p.HealthSecond = defaultValue(p.HealthSecond, 5)
	p.IdleTime = defaultValue(p.IdleTime, 60)
	if p.MinPoolSize > p.MaxPoolSize {
		p.MinPoolSize = p.MaxPoolSize
	}
}

// 获取默认值
//
//  param，int，参数值
//  defaultValue，int，默认返回
//  返回，int。如果参数值小于1就返回默认值，否则返回参数值。
func defaultValue(param, defaultValue int) int {
	if param < 1 {
		return defaultValue
	} else {
		return param
	}
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (p *Pool) Start() error {
	p.defaultConfig()
	p.poolWait = make(chan *PooledClient, p.MaxWaitSize)
	p.waitCount = 0
	p.pooled.Init(p.AcquireIncrement, p.MinPoolSize, p.MaxPoolSize, p)
	err := p.pooled.Append(p.MinPoolSize)
	if err != nil {
		return err
	}
	p.Status = PoolStart

	p.watcher = time.NewTicker(time.Second * time.Duration(p.HealthSecond))
	go p.watch()
	return nil
}

//创建新的连接池
//
// 返回 Pool
func NewPool() *Pool {
	return &Pool{
		pooled: new(Slice),
	}
}

//返回连接池的状态信息
//
// 返回，string
func (p *Pool) Info() string {
	return fmt.Sprintf(`pool size:%d	actived client:%d	wait create:%d	config max pool size:%d	`,
		p.pooled.length, p.pooled.current, p.waitCount, p.MaxPoolSize)
}

//关闭连接池
func (p *Pool) Close() {
	if p.watcher != nil {
		p.watcher.Stop()
	}
	p.Status = PoolStop
	close(p.poolWait)
	p.pooled.Close()
}
func (p *Pool) checkWait() error {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.waitCount >= p.MaxWaitSize {
		return goerr.New("pool is busy,Wait for connection creation has reached %d", p.waitCount)
	}
	return nil
}

//在连接池取一个新连接
//
//  返回 client，一个新的连接
//  返回 err，可能的错误，操作成功返回 nil
func (p *Pool) Get() (client *PooledClient, err error) {
	switch p.Status {
	case PoolStop:
		return nil, goerr.New("the Connectors is Closed, can not get new client.")
	case PoolInit:
		return nil, goerr.New("the Connectors is not initialized, can not get new client.")
	}
	//检查是否有缓存的连接
	client, err = p.pooled.Get()
	if err == nil {
		return
	}
	//检查是否可以扩展
	if err = p.pooled.Append(p.AcquireIncrement); err == nil {
		client, err = p.pooled.Get()
		if err == nil {
			return
		}
	}
	if err = p.checkWait(); err != nil {
		return nil, err
	}
	p.lock.Lock()
	p.waitCount += 1
	p.lock.Unlock()
	//enter slow poolWait
	timeout := time.After(time.Duration(p.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		err = goerr.New("pool is busy,can not get new client in %d seconds", p.GetClientTimeout)
	case cc := <-p.poolWait:
		if cc == nil {
			err = goerr.New("pool is Closed, can not get new client.")
		} else {
			client = cc
			err = nil
		}
	}
	p.lock.Lock()
	p.waitCount -= 1
	p.lock.Unlock()
	return
}

//归还连接到连接池
//
//  element 连接
func (p *Pool) Set(element *PooledClient) {
	if element == nil {
		return
	}
	if p.Status == PoolStart {
		element.lastTime = now //设置最好的回收时间
		p.lock.RLock()
		defer p.lock.RUnlock()
		if p.waitCount > 0 && element.Client.IsOpen() && element.index%5 == 0 {
			p.poolWait <- element
		} else {
			p.pooled.setPoolClient(element)
		}
	} else {
		if element.Client.IsOpen() {
			element.Client.Close()
		}
	}
}
