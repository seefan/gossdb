package main

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"time"

	log "github.com/cihub/seelog"
	//"net/http"
	_ "net/http/pprof"
)

func main() {
	if logger, err := log.LoggerFromConfigAsFile("./log.xml"); err == nil {
		log.ReplaceLogger(logger)
	}
	//go func() {
	//	http.ListenAndServe("0.0.0.0:6060", nil) // 启动默认的 http 服务，可以使用自带的路由
	//}()
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		MaxWaitSize:      0,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Critical("create pool error", err)
	}
	defer pool.Close()
	for i := 0; i < 2; i++ {
		go func(idx int) {
			for {
				Test_hset1(pool, idx)
			}
		}(i)
	}
	time.Sleep(time.Hour * 24)
}
func Test_hset1(pool *gossdb.Connectors, i int) {

	c, err := pool.NewClient()
	if err != nil {
		log.Error("create", i, err, pool.Info())
		return
	}
	defer c.Close()

	err = c.Hset("hset", "test", "hello world.")
	if err != nil {
		log.Error(i, err)
	} else {
		log.Info("is set", i)
	}
	re, err := c.Hget("hset", "test")
	if err != nil {
		log.Error(i, err)
	} else {
		log.Info(re, "is get", i)
	}
	log.Info(pool.Info())
	md := make(map[string]interface{})
	md["abc"] = "abc1"
	md["ab"] = "abc"
	err = c.MultiHset("hset", md)
	if err != nil {
		log.Error(i, err)
	} else {
		log.Info("is mhset", i)
	}
	m, err := c.MultiHget("hset", "ab", "test1")
	if err != nil {
		log.Error(i, err)
	} else {
		log.Info(m, "is mhget", i)
	}
	time.Sleep(time.Millisecond)
	//c.Close()
}
