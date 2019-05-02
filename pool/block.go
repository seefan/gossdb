/*
@Time : 2019-04-30 20:26
@Author : seefan
@File : block
@Software: gossdb
*/
package pool

import (
	"sync"
	"time"

	"github.com/seefan/goerr"
)

// 使用slice实现的池
type Block struct {
	//id
	id int32
	//长度
	length int
	//lock
	lock sync.Mutex
	//element list
	pooled []*PooledClient
	//available
	available bool
	// last error
	lastError error
	//ava id
	queue *Queue
}

//附加连接
//
// ai int 附加的连接数
// 返回 error 可能的错误
func newBlock(cs []*PooledClient) *Block {
	s := new(Block)
	s.length = len(cs)
	s.pooled = make([]*PooledClient, s.length)

	s.queue = newQueue(s.length)
	for i := 0; i < s.length; i++ {
		s.pooled[i] = cs[i]
		s.pooled[i].index = i
		s.pooled[i].pool = s
		s.queue.Put(i)
	}
	return s
}

//获取一个连接
//
// 返回 *PooledClient 缓存的连接
// 返回 error 可能的错误
func (s *Block) Get() (*PooledClient, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	idx := s.queue.Pop()
	if idx == -1 {
		return nil, goerr.String("pool is empty")
	}
	return s.pooled[idx], nil
}

//回收连接
//
// element *PooledClient 要回收的连接
func (s *Block) Set(element *PooledClient) {
	if element.Client.IsOpen() {
		s.setPoolClient(element)
	} else {
		s.resetClient(element)
	}
}

//回收连接
func (s *Block) setPoolClient(element *PooledClient) {
	s.lock.Lock()
	defer s.lock.Unlock()
	element.isUsed = false
	element.lastTime = time.Now().Unix()
	s.queue.Put(element.index)
}

//重置连接，重启连接，如果不成功，则将整个块设置为不可用
//
// element *PooledClient 要回收的连接
func (s *Block) resetClient(element *PooledClient) {
	err := element.Client.Close()
	if err != nil {
		s.lastError = err
	}
	err = element.Client.Start()
	if err != nil {
		s.lastError = err
		s.available = false
	} else {
		s.setPoolClient(element)
	}
}

//关闭连接池
func (s *Block) Close() {
	for _, c := range s.pooled {
		if c != nil && c.Client.IsOpen() {
			_ = c.Client.Close()
		}
	}
	s.pooled = s.pooled[:0]
}
