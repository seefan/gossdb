package gopool

import (
	"testing"
	//"time"
)

type SSDBClient struct {
	isOpen bool
}

//打开连接
func (s *SSDBClient) Start() error {
	s.isOpen = true
	return nil
}
func (s *SSDBClient) Close() error {
	s.isOpen = false
	return nil
}
func (s *SSDBClient) IsOpen() bool {
	return s.isOpen
}
func (s *SSDBClient) Ping() bool {
	return s.Start() == nil
}

func BenchmarkGetSet(b *testing.B) {
	pool := NewPool()
	pool.NewClient = func() IClient {
		return &SSDBClient{}
	}

	pool.MinPoolSize = 10
	pool.MaxPoolSize = 200
	pool.MaxWaitSize = 100000
	pool.GetClientTimeout = 5
	pool.HealthSecond = 10
	if err := pool.Start(); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		if c, e := pool.Get(); e == nil {
			pool.Set(c)
		} else {
			b.Error(e)
		}
	}
}

func BenchmarkP(b *testing.B) {
	pool := NewPool()
	pool.NewClient = func() IClient {
		return &SSDBClient{}
	}
	pool.MinPoolSize = 10
	pool.MaxPoolSize = 200
	pool.MaxWaitSize = 100000
	pool.GetClientTimeout = 5
	pool.HealthSecond = 10
	if err := pool.Start(); err != nil {
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if c, e := pool.Get(); e == nil {
				//time.Sleep(time.Millisecond)
				pool.Set(c)
			} else {
				b.Error(e)
			}
		}
	})
}
