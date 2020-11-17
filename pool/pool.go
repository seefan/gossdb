package pool

import (
	"sync"

	"github.com/seefan/gossdb/consts"
	"github.com/seefan/gossdb/queue"
)

//Pool pool block
// 连接池结构
type Pool struct {
	index int32 //pos
	//连接数
	size int
	//element list
	pooled []*Client
	//available index
	available Avaliable
	//状态
	status int32
	//new client
	New func() (*Client, error)
	//lock
	lock sync.Mutex
	//health 0 正常 1 检查 2 关闭中
	health int32
}

//新建一个池
func newPool(size int) *Pool {
	return &Pool{
		pooled:    make([]*Client, size),
		size:      size,
		available: queue.NewQueue(size),
		//available: newRing(size),
		status: consts.None,
	}
}

// //CheckClose 检查是否可以关闭
func (p *Pool) CheckClose() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.status == consts.PoolStop {
		p.Close()
	}
	if p.available.IsEmpty() && p.status != consts.PoolStop {
		p.status = consts.PoolStop
	}
}

//CheckHeath check opened number
func (p *Pool) CheckHeath() {
	p.lock.Lock()
	defer p.lock.Unlock()
	count := 0
	for _, c := range p.pooled {
		if c.IsOpen() {
			count++
		}
	}
	if count == p.size {
		p.status = consts.PoolStart
	}
}

//Start start the pool
//
//  @return error possible error, operation successfully returned nil
//
//启动连接
func (p *Pool) Start() (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for i := 0; i < p.size; i++ {
		c := p.pooled[i]
		if c == nil {
			cc, err := p.New()
			if err != nil {
				return err
			}
			cc.index = i
			p.pooled[i] = cc
		}
		if !p.pooled[i].IsOpen() {
			if err := p.pooled[i].Start(); err != nil {
				return err
			}
		}
	}
	p.status = consts.PoolStart
	return nil
}

//Close close pool
//
//关闭连接池
func (p *Pool) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.status = consts.PoolStop
	for _, c := range p.pooled {
		if c != nil {
			_ = c.SSDBClient.Close()
		}
	}
}

//Get get a pooled connection
//
//  @return *Client，client
//  @return error possible error, operation successfully returned nil
//
//获取一个缓存的连接
func (p *Pool) Get() (client *Client) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.status == consts.PoolStop {
		return nil
	}
	pos := p.available.Pop()
	if pos == -1 {
		return nil
	}
	return p.pooled[pos]
}

//Set return the connection to the connection pool
//
//  @param client thd connection
//
//返还一个连接到连接池
func (p *Pool) Set(client *Client) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if client == nil {
		return
	}
	p.available.Put(client.index)
	if p.status == consts.PoolStop {
		if client.IsOpen() {
			_ = client.SSDBClient.Close()
		}
	}
}
