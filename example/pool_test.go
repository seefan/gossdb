package main

import (
	log "github.com/cihub/seelog"
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"testing"
)

func BenchmarkGetSet(b *testing.B) {
	if logger, err := log.LoggerFromConfigAsFile("./log.xml"); err == nil {
		log.ReplaceLogger(logger)
	}
	defer log.Flush()
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      10,
		MaxPoolSize:      100,
		MaxWaitSize:      10000,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Critical(err)
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
		log.Critical(err)
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
