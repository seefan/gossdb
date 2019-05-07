/*
@Time : 2019-05-06 19:32
@Author : seefan
@File : pool
@Software: gossdb
*/
package pool

import (
	"errors"
	"sync"
)

// 连接池结构
type Pool struct {
	index int32 //pos
	//连接数
	size int
	//element list
	pooled []*Client
	//available index
	available *Queue
	//状态
	Status int
	//lock
	lock sync.Mutex
	//new client
	New func() (*Client, error)
}

func newPool(size int) *Pool {
	return &Pool{
		pooled:    make([]*Client, size),
		size:      size,
		available: newQueue(size),
	}
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (p *Pool) Start() error {
	if p.Status != PoolInit {
		return errors.New("pool status not init")
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	for i := 0; i < p.size; i++ {
		if c, err := p.New(); err == nil {
			c.index = i
			p.pooled[i] = c
			p.available.Add(i)
		} else {
			return err
		}
	}
	p.Status = PoolStart
	return nil
}

//关闭连接池
func (p *Pool) Close() {
	p.Status = PoolStop
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
func (p *Pool) Get() (client *Client, err error) {
	if p.Status != PoolStart {
		return nil, errors.New("pool is not start")
	}
	//检查是否有缓存的连接
	p.lock.Lock()
	defer p.lock.Unlock()
	pos := p.available.Pop()
	if pos == -1 {
		return nil, errors.New("pool is empty")
	}
	return p.pooled[pos], nil
}

//归还连接到连接池
//
//  element 连接
func (p *Pool) Set(element *Client) {
	if element == nil {
		return
	}
	if p.Status == PoolStart {
		p.lock.Lock()
		pos := p.available.Put(element.index)
		p.lock.Unlock()
		if pos < 0 { //put failed
			_ = element.SSDBClient.Close()
		}
	} else {
		_ = element.SSDBClient.Close()
	}
}
