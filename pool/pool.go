package pool

import (
	"errors"

	"github.com/seefan/gossdb/consts"
)

// 连接池结构
type Pool struct {
	index byte //pos
	//连接数
	size int
	//element list
	pooled []*Client
	//available index
	available *Queue
	//状态
	status int
	//new client
	New func() (*Client, error)
}

func newPool(size int) *Pool {
	return &Pool{
		pooled:    make([]*Client, size),
		size:      size,
		available: newQueue(size),
		status:    consts.PoolStop,
	}
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (p *Pool) Start() error {
	if p.status != consts.PoolStop {
		return errors.New("pool already start")
	}

	for i := 0; i < p.size; i++ {
		if c, err := p.New(); err == nil {
			c.index = i
			p.pooled[i] = c
		} else {
			return err
		}
	}
	p.status = consts.PoolStart
	return nil
}

//关闭连接池
func (p *Pool) Close() {
	p.status = consts.PoolStop
	for _, c := range p.pooled {
		if c != nil {
			_ = c.SSDBClient.Close()
		}
	}
}

//在连接池取一个新连接
//
//  返回 client，一个新的连接
//  返回 err，可能的错误，操作成功返回 nil
func (p *Pool) Get() (client *Client) {
	if p.status == consts.PoolNotStart {
		return nil
	}
	//检查是否有缓存的连接
	pos := p.available.Pop()
	if pos == -1 {
		return nil
	}
	return p.pooled[pos]
}

//归还连接到连接池
//
//  element 连接
func (p *Pool) Set(client *Client) {
	if client == nil {
		return
	}
	if p.status == consts.PoolStop {
		if client.IsOpen() {
			_ = client.SSDBClient.Close()
		}
	} else {
		p.available.Put(client.index)
	}
}
