package gossdb

import (
	"fmt"
	"github.com/ssdb/gossdb/ssdb"
	//	"log"
	//	"log"
	"sync"
	"time"
)

//连接池
type Connectors struct {
	pool        chan *Client //连接池
	cfg         *Config      //配置
	lock        sync.RWMutex //锁
	Size        int          //连接池大小
	ActiveCount int          //活动连接数
	WaitCount   int          //当前等待创建的连接数
	Status      int          //状态 0：创建 1：正常 -1：关闭
	Encoding    bool         //是否启动编码，启用后会对struct 进行 json 编码，以支持更多类型
}

const (
	//连接池状态：创建
	PoolInit = 0
	//连接池状态：正常
	PoolStart = 1
	//连接池状态：关闭
	PoolStop = -1
)

//根据配置初始化连接池
//
//  conf 连接池的初始化配置
//
//默认值
//
//	GetClientTimeout int 获取连接超时时间，单位为秒，默认1分钟
//	MaxPoolSize int 最大连接池个数，默认为10
//	MinPoolSize int 最小连接池数，默认为1
//	AcquireIncrement int  当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 3
//	MaxIdleTime int 最大空闲时间，指定秒内未使用则连接被丢弃。若为0则永不丢弃。默认值: 0
//  MaxWaitSize int 最大等待数目，当连接池满后，新建连接将排除等待池中连接释放，本值限制最大等待的数量。默认值: 1000
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
	if conf.MaxWaitSize < 1 {
		conf.MaxWaitSize = 1000
	}
	if conf.MinPoolSize > conf.MaxPoolSize {
		conf.MinPoolSize = conf.MaxPoolSize
	}
	c := &Connectors{
		pool: make(chan *Client, conf.MaxPoolSize),
		cfg:  conf,
	}
	for i := 0; i < conf.MinPoolSize; i++ {
		if nc, err := c.newClient(); err == nil {
			c.pool <- nc
			c.Size += 1
		} else {
			return nil, err
		}
	}

	if c.Size == 0 {
		return nil, fmt.Errorf("create pool is failed.")
	}
	c.Status = PoolStart
	if c.cfg.MaxIdleTime > 0 {
		go c.timed()
	}
	return c, nil
}

//设置连接池的活动连接变化
func (this *Connectors) SetCount(active, wait, size int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if active != 0 {
		this.ActiveCount += active
	}
	if wait != 0 {
		this.WaitCount += wait
	}
	if size != 0 {
		this.Size += size
	}
}

//关闭连接池
func (this *Connectors) Close() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.Status = PoolStop
	for len(this.pool) > 0 {
		c := <-this.pool
		c.Client.Close()
	}
	close(this.pool)
}

//定期执行，简单方式
func (this *Connectors) timed() {
	timer := time.Tick(time.Minute)
	for t := range timer {
		go this.Contraction(t)
	}
}

//收缩连接池
//
//  now 当前的时间
func (this *Connectors) Contraction(now time.Time) {
	this.lock.Lock()
	defer this.lock.Unlock()
	//只有正常运行时处理
	if this.Status != PoolStart {
		return
	}
	//太忙时不处理
	if this.WaitCount > 0 {
		return
	}
	tmplist := []*Client{}
	for len(this.pool) > 0 {
		c := <-this.pool
		if c.lastTime.Add(time.Duration(this.cfg.MaxIdleTime) * time.Second).After(now) {
			tmplist = append(tmplist, c)
		} else {
			c.Client.Close()
		}
	}
	for _, c := range tmplist {
		this.pool <- c
	}
	this.Size = len(this.pool)
}

//状态信息
func (this *Connectors) Info() string {
	return fmt.Sprintf(`pool size:%d,%d	actived client:%d	wait create:%d	config max pool size:%d	config Increment:%d`,
		this.Size, len(this.pool), this.ActiveCount, this.WaitCount, this.cfg.MaxPoolSize, this.cfg.AcquireIncrement)
}

//从连接池里获取一个新连接
//
//  返回 一个新连接
//  返回 可能的错误
func (this *Connectors) NewClient() (*Client, error) {
	if err := this.checkNew(); err != nil {
		return nil, err
	}
	this.SetCount(0, 1, 0)
	//此处如果无法增加连接，会导致在超时的时间内，所有的连接都无法释放
	timeout := time.After(time.Duration(this.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		return nil, fmt.Errorf("ssdb pool is busy,can not get new client in %d seconds", this.cfg.GetClientTimeout)
	case c := <-this.pool:
		this.SetCount(1, -1, -1)
		return c, nil
	}
}

//检查是否有新连接可以创建
func (this *Connectors) checkNew() error {
	this.lock.Lock()
	defer this.lock.Unlock()
	switch this.Status {
	case -1:
		return fmt.Errorf("the Connectors is Closed, can not get new client.")
	case 0:
		return fmt.Errorf("the Connectors is not inited, can not get new client.")
	}
	if this.WaitCount >= this.cfg.MaxWaitSize {
		return fmt.Errorf("ssdb pool is busy,Wait for connection creation has reached %d", this.WaitCount)
	}
	if this.Size+this.ActiveCount < this.cfg.MaxPoolSize { //如果没有连接了，检查是否可以自动增加
		for i := 0; i < this.cfg.AcquireIncrement && this.Size+this.ActiveCount < this.cfg.MaxPoolSize; i++ {
			if c, err := this.newClient(); err == nil {
				this.Size += 1
				this.pool <- c
			} else { //如果新建连接有错误，就放弃
				break
			}
		}
	}
	return nil
}

//关闭一个连接，如果连接池关闭，就销毁连接，否则就回收到连接池
//
//  client 要关闭的连接
func (this *Connectors) closeClient(client *Client) {
	if this.isStart() {
		client.lastTime = time.Now()
		this.pool <- client
		this.SetCount(-1, 0, 1)
	} else {
		client.Client.Close()
	}
}

//检查连接池是否是开启状态
func (this *Connectors) isStart() bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.Status == PoolStart
}

//创建一个连接
//
//  返回 一个新连接
//  返回 可能的错误
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
