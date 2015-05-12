package gossdb

import (
	"fmt"
	"github.com/ssdb/gossdb/ssdb"
	//	"log"
	"sync"
	"time"
)

//连接池
type Connectors struct {
	pool        chan *Client //连接池
	cfg         *Config      //配置
	lock        sync.Mutex   //锁
	Size        int          //连接池大小
	ActiveCount int          //活动连接数
	Status      int          //状态 0：创建 1：正常 -1：关闭
	Encoding    bool         //是否启动编码，启用后会对struct 进行 json 编码，以支持更多类型
}

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

	if c.Size == 0 {
		return nil, fmt.Errorf("create pool is failed.")
	}
	c.Status = 1
	if c.cfg.MaxIdleTime > 0 {
		go c.timed()
	}
	return c, nil
}

//关闭连接池
func (this *Connectors) Close() {
	this.Status = -1
	for len(this.pool) > 0 {
		c := <-this.pool
		c.Client.Close()
	}
	this.ActiveCount = 0
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
	if this.Status != 1 {
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
	return fmt.Sprintf(`pool size:%d	actived client:%d	config max pool size:%d	config Increment:%d`, this.Size, this.ActiveCount, this.cfg.MaxPoolSize, this.cfg.AcquireIncrement)
}

//从连接池里获取一个新连接
//
//  返回 一个新连接
//  返回 可能的错误
func (this *Connectors) NewClient() (*Client, error) {
	this.lock.Lock()
	switch this.Status {
	case -1:
		return nil, fmt.Errorf("the Connectors is Closed, can not get new client.")
	case 0:
		return nil, fmt.Errorf("the Connectors is not inited, can not get new client.")
	}
	if this.Size <= this.ActiveCount && this.Size < this.cfg.MaxPoolSize { //如果没有连接了，检查是否可以自动增加
		for i := 0; i < this.cfg.AcquireIncrement && this.Size < this.cfg.MaxPoolSize; i++ {
			if c, err := this.newClient(); err == nil {
				this.pool <- c
				this.Size = len(this.pool) + this.ActiveCount
			} else { //如果新建连接有错误，就放弃
				break
			}
		}
	}
	//此处如果无法增加连接，会导致在超时的时间内，所有的连接都无法释放
	this.lock.Unlock()
	//首先把锁释放
	now := time.Now()
	timeoutSeconds := float64(this.cfg.GetClientTimeout)
	var waitClient *Client
	waitTimeout := true
	for {
		timeout := time.After(time.Millisecond) //千分之一秒超时
		this.lock.Lock()
		select {
		case <-timeout:
			//检查等待时间是否超时
			if time.Since(now).Seconds() > timeoutSeconds {
				waitTimeout = false
			}
		case c := <-this.pool:
			this.ActiveCount += 1
			waitClient = c
			waitTimeout = false
		}
		this.lock.Unlock()
		if waitTimeout {
			time.Sleep(time.Millisecond)
		} else {
			break
		}

	}
	if waitClient != nil {
		return waitClient, nil
	} else {
		return nil, fmt.Errorf("ssdb pool is busy,can not get new client in %d seconds", this.cfg.GetClientTimeout)
	}
}

//关闭一个连接，如果连接池关闭，就销毁连接，否则就回收到连接池
//
//  client 要关闭的连接
func (this *Connectors) closeClient(client *Client) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.Status != 1 {
		client.Client.Close()
	} else {
		client.lastTime = time.Now()
		this.pool <- client
		this.ActiveCount -= 1
	}
}

//按要求创建连接
//
//  size 一次创建多少个连接
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
	this.Size = len(this.pool)
	return nil
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
