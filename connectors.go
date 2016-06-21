package gossdb

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
	//追加连接
	PoolAppend = 1
	//减少连接
	PoolReduct = -1
)

//连接池
type Connectors struct {
	pool        chan *Client //连接池
	poolMap     map[*Client]bool
	cfg         *Config      //配置
	lock        sync.Mutex   //全局锁
	WaitCount   int          //当前等待创建的连接数
	Status      int          //状态 0：创建 1：正常 -1：关闭
	timeHealth  *time.Ticker //健康检查控制
	index       int          //位置记录
	Size        int          //连接池大小
	ActiveCount int          //活动连接数
	start       int          //当前连接池id
	end         int          //回收连接池id
}

//用配置文件进行初始化
//
//  cfg 配置文件
func (this *Connectors) Init(cfg *Config) {
	this.setConfig(cfg)
	this.pool = make(chan *Client, cfg.MaxPoolSize)
	this.poolMap = make(map[*Client]bool)
	this.timeHealth = time.NewTicker(time.Second)
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (this *Connectors) Start() error {
	this.Size = 0
	this.ActiveCount = 0
	this.WaitCount = 0
	for i := 0; i < this.cfg.MinPoolSize; i++ {
		c := this.createClient()
		if err := c.Start(); err != nil {
			return goerr.NewError(err, "启动连接池出错")
		}
		this.Size += 1
		this.pool <- c
		this.poolMap[c] = false
	}
	this.Status = PoolStart
	go func() {
		keep := 0
		for range this.timeHealth.C {
			//db check
			if keep%this.cfg.DBHealthSecond == 0 {
				c, err := this.NewClient()
				if err != nil {
					keep = this.cfg.HealthSecond
				} else {
					if !c.Ping() {
						keep = this.cfg.HealthSecond
					}
					c.Close()
				}
			}
			keep += 1
			if keep >= this.cfg.HealthSecond {
				this.healthWorker()
				keep = 0
			}
		}
	}()
	return nil
}

//设置配置文件，主要是设置默认值
//
//  cfg 配置文件
func (this *Connectors) setConfig(conf *Config) {
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
		conf.HealthSecond = 300
	}
	if conf.DBHealthSecond < 1 {
		conf.DBHealthSecond = 5
	}
	if conf.MinPoolSize > conf.MaxPoolSize {
		conf.MinPoolSize = conf.MaxPoolSize
	}
	this.cfg = conf
}

//连接池健康检查
//
//主要检查连接是否过期，是否可用，只设置状态，不处理
func (this *Connectors) healthWorker() {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.WaitCount == 0 {
		//fmt.Println("heath check", this.Info())
		//是否需要检查过期
		size := this.Size
		expired := 0
		for i := 0; i < size; i++ {
			c := <-this.pool
			this.Size -= 1
			//先检查过期，再检查是否连通
			if this.cfg.MaxIdleTime > 0 && this.Size >= this.cfg.MinPoolSize && expired < this.cfg.AcquireIncrement && time.Since(c.lastTime).Seconds() > float64(this.cfg.MaxIdleTime) { //过期
				c.Client.Close()
				delete(this.poolMap, c)
				expired += 1
			} else {
				if c.Ping() { //正常通信，直接回收
					this.pool <- c
					this.Size += 1
				} else {
					c.Client.Close() //重启一下
					if err := c.Start(); err == nil {
						this.pool <- c
						this.Size += 1
					} else {
						delete(this.poolMap, c)
					}
				}
			}
		}
	}

}

//归还连接到连接池
//
//  c 连接
func (this *Connectors) closeClient(c *Client) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if c == nil {
		return
	}
	this.ActiveCount -= 1
	this.Size += 1
	if this.Status == PoolStart {
		this.poolMap[c] = false
		this.pool <- c
	} else {
		c.Client.Close()
		if this.poolMap != nil {
			delete(this.poolMap, c)
		}
	}
}

//在连接池取一个新连接
//
//  返回 client，一个新的连接
//  返回 err，可能的错误，操作成功返回 nil
func (this *Connectors) NewClient() (*Client, error) {
	if err := this.checkNew(); err != nil {
		return nil, err
	}
	this.lock.Lock()
	this.WaitCount += 1
	this.lock.Unlock()
	timeout := time.After(time.Duration(this.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		return nil, goerr.New("ssdb pool is busy,can not get new client in %d seconds", this.cfg.GetClientTimeout)
	case c := <-this.pool:
		if c == nil {
			return nil, goerr.New("the Connectors is Closed, can not get new client.")
		}
		this.lock.Lock()
		this.WaitCount -= 1
		this.ActiveCount += 1
		this.Size -= 1
		this.poolMap[c] = true
		this.lock.Unlock()
		return c, nil
	}

}

//检查是否可以创建新连接，是否需要增加连接数
//
//  返回 err，可能的错误，操作成功返回 nil
func (this *Connectors) checkNew() error {
	this.lock.Lock()
	defer this.lock.Unlock()
	switch this.Status {
	case PoolStop:
		return goerr.New("the Connectors is Closed, can not get new client.")
	case PoolInit:
		return goerr.New("the Connectors is not inited, can not get new client.")
	}
	if this.WaitCount >= this.cfg.MaxWaitSize {
		return goerr.New("ssdb pool is busy,Wait for connection creation has reached %d", this.WaitCount)
	}
	if this.Size < this.cfg.MinPoolSize && this.Size+this.ActiveCount < this.cfg.MaxPoolSize { //如果没有连接了，检查是否可以自动增加
		for i := 0; i < this.cfg.AcquireIncrement; i++ {
			c := this.createClient()
			if err := c.Start(); err == nil {
				this.pool <- c
				this.poolMap[c] = false
				this.Size += 1
			} else {
				return err
			}
		}
	}
	return nil
}

//关闭连接池
func (this *Connectors) Close() {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.Status == PoolStart {
		close(this.pool)
	}
	this.Status = PoolStop
	//关闭连接池中的连接
	for p, isRun := range this.poolMap {
		if !isRun {
			p.Client.Close()
			this.Size -= 1
			delete(this.poolMap, p)
		}
	}
	this.timeHealth.Stop()
}

//状态信息
//
//  返回 string，一个详细连接池基本情况的字符串
func (this *Connectors) Info() string {
	return fmt.Sprintf(`pool size:%d	actived client:%d	wait create:%d	config max pool size:%d	config Increment:%d`,
		this.Size, this.ActiveCount, this.WaitCount, this.cfg.MaxPoolSize, this.cfg.AcquireIncrement)
}

//创建一个新的连接
func (this *Connectors) createClient() *Client {
	c := new(Client)
	c.pool = this
	return c
}
