package main

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"testing"
)

func BenchmarkGetSet(b *testing.B) {

	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      10,
		MaxPoolSize:      100,
		MaxWaitSize:      10000,
		AcquireIncrement: 5,
	})
	if err != nil {
		b.Fatal("create pool error", err)
	}
	defer pool.Close()
	for i := 0; i < b.N; i++ {
		if c, e := pool.NewClient(); e == nil {
			c.Close()
		} else {
			b.Error(e)
		}

	}
}

func BenchmarkP(b *testing.B) {
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      10,
		MaxPoolSize:      100,
		MaxWaitSize:      10000,
		AcquireIncrement: 5,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer pool.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if c, e := pool.NewClient(); e == nil {
				c.Close()
			} else {
				b.Error(e)
			}
		}
	})
}
