package gossdb

import (
	"fmt"
	"github.com/ssdb/gossdb/ssdb"
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
	defer this.lock.Unlock()
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
	timeout := time.Tick(time.Duration(this.cfg.GetClientTimeout) * time.Second)
	select {
	case <-timeout:
		return nil, fmt.Errorf("ssdb pool is busy,can not get new client")
	case c := <-this.pool:
		this.ActiveCount += 1
		return c, nil
	}
}

//关闭一个连接，如果连接池关闭，就销毁连接，否则就回收到连接池
//
//client 要关闭的连接
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
//size 创建多少个
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
