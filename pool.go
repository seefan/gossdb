/*
@Time : 2019-05-06 19:32
@Author : seefan
@File : pool
@Software: gossdb
*/
package gossdb

import (
	"errors"
	"sync"

	"github.com/seefan/gossdb/consts"
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
		Status:    consts.PoolStop,
	}
}

//启动连接池
//
//  返回 err，可能的错误，操作成功返回 nil
func (p *Pool) Start() error {
	if p.Status != consts.PoolStop {
		return errors.New("pool already start")
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
	p.Status = consts.PoolStart
	return nil
}

//关闭连接池
func (p *Pool) Close() {
	p.Status = consts.PoolStop
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
	if p.Status == consts.PoolNotStart {
		return nil
	}
	//检查是否有缓存的连接
	p.lock.Lock()
	pos := p.available.Pop()
	p.lock.Unlock()
	if pos == -1 {
		return nil
	}
	return p.pooled[pos]
}

//归还连接到连接池
//
//  element 连接
func (p *Pool) Set(element *Client) {
	if element == nil {
		return
	}
	if p.Status == consts.PoolStop {
		if element.IsOpen() {
			_ = element.SSDBClient.Close()
		}
	} else {
		p.lock.Lock()
		p.available.Put(element.index)
		p.lock.Unlock()
	}
}
