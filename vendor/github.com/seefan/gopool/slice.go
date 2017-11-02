package gopool

import (
	"sync"
	"github.com/seefan/goerr"
	"fmt"
)

// 使用slice实现的池
type Slice struct {
	//记数
	//可用长度
	length int
	//可用连接位置
	current int
	//lock
	//lockCurrent sync.RWMutex
	lock sync.Mutex
	//最大连接池个数。默认值: 20
	maxPoolSize int
	//最小连接池数。默认值: 5
	minPoolSize int
	//当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
	acquireIncrement int
	//element list
	pooled []*PooledClient
	//连接池
	pool *Pool
}

//初始化连接池
//
// ai int acquireIncrement
// min int minPoolSize
// max int maxPoolSize
// p *Pool 主连接池
func (s *Slice) Init(ai, min, max int, p *Pool) {
	s.acquireIncrement = ai
	s.minPoolSize = min
	s.maxPoolSize = max
	s.pool = p
	s.pooled = make([]*PooledClient, max)
	s.length = 0
	s.current = 0
}

//附加连接
//
// ai int 附加的连接数
// 返回 error 可能的错误
func (s *Slice) Append(ai int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.length < s.maxPoolSize {
		size := s.length
		for i := 0; i < ai && size < s.maxPoolSize; i++ {
			client := s.pooled[size]
			if client == nil {
				client = s.newPooledClient()
			}
			if client.Client.IsOpen() == false {
				if err := client.Client.Start(); err != nil {
					return goerr.NewError(err, "can not create client")
				}
			}
			s.pooled[size] = client
			client.index = size
			size += 1
		}
		s.length = size
	}
	return nil
}

//创建新的缓存连接
func (s *Slice) newPooledClient() *PooledClient {
	return &PooledClient{
		Client: s.pool.NewClient(),
		pool:   s.pool,
	}
}

func (s *Slice) Info(v ...interface{}) {
	b := []int{}
	for _, v := range s.pooled {
		if v != nil {
			b = append(b, v.index)
		}
	}
	fmt.Println(v, b)
}

//获取一个连接
//
// 返回 *PooledClient 缓存的连接
// 返回 error 可能的错误
func (s *Slice) Get() (*PooledClient, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.current < s.length {
		element := s.pooled[s.current]
		element.isUsed = true
		element.index = s.current
		s.current += 1
		return element, nil
	}
	return nil, goerr.New("pool is empty")
}

//回收连接
//
// element *PooledClient 要回收的连接
func (s *Slice) Set(element *PooledClient) {
	if element.Client.IsOpen() {
		s.setPoolClient(element)
	} else {
		s.closeClient(element)
	}
}

//回收连接
func (s *Slice) setPoolClient(element *PooledClient) {
	s.lock.Lock()
	defer s.lock.Unlock()
	element.isUsed = false
	pos := s.current - 1
	idx := element.index
	if element.index < pos {
		s.pooled[pos].index, element.index = element.index, s.pooled[pos].index
		s.pooled[pos], s.pooled[idx] = s.pooled[idx], s.pooled[pos]
	}
	s.current -= 1
}

//关闭连接
func (s *Slice) closeClient(element *PooledClient) {
	s.lock.Lock()
	defer s.lock.Unlock()
	element.isUsed = false
	pos := s.length - 1
	if element.index < pos {
		s.pooled[pos], s.pooled[element.index] = s.pooled[element.index], s.pooled[pos]
	}
	s.length -= 1
	if s.current == pos {
		s.current -= 1
	}
}

//关闭连接池
func (s *Slice) Close() {
	size := len(s.pooled)
	for i := 0; i < size; i++ {
		if c := s.pooled[i]; c != nil && c.Client.IsOpen() {
			c.Client.Close()
		}
	}
}
