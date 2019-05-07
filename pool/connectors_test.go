/*
@Time : 2019-05-07 15:09
@Author : seefan
@File : connectors_test.go
@Software: gossdb
*/
package pool

import (
	"sync"
	"testing"

	"github.com/seefan/gossdb/conf"
)

func TestConnectors_NewClient(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    100,
		PoolNumber:  5,
	})
	err := pool.Start()
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 10; i++ {
		c, err := pool.NewClient()
		if err == nil {
			t.Log(c.Info())
			c.Close()
		} else {
			t.Error(err)
		}
	}
	pool.Close()
}
func BenchmarkConnectors_NewClient(b *testing.B) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    20,
		PoolNumber:  20,
	})
	err := pool.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c, err := pool.NewClient()
			if err == nil {
				//_, _ = c.Info()
				c.Close()
			} else {
				//b.Error(err)
			}
		}
	})

	pool.Close()
}
func Test1000(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    50,
		PoolNumber:  20,
	})
	err := pool.Start()
	if err != nil {
		panic(err)
	}
	var wait sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				c, err := pool.NewClient()
				if err == nil {
					_, _ = c.Get("a")
					c.Close()
				} else {
					t.Error(err)
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
	pool.Close()
}
