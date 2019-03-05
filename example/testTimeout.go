package main

import (
	//"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

func main() {

	//go func() {
	//	http.ListenAndServe("0.0.0.0:6060", nil) // 启动默认的 http 服务，可以使用自带的路由
	//}()
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      1,
		MaxPoolSize:      10,
		MaxWaitSize:      100,
		AcquireIncrement: 5,
	})
	if err != nil {
		panic(err.Error())
	}
	defer pool.Close()
	for i := 0; i < 20; i++ {
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
		println("create", i, err, pool.Info())
		return
	}
	defer c.Close()

	err = c.Hset("hset", "test", "hello world.")
	if err != nil {
		println(i, err)
	} else {
		println("is set", i)
	}
	re, err := c.Hget("hset", "test")
	if err != nil {
		println(i, err)
	} else {
		println(re, "is get", i)
	}
	println(pool.Info())
	md := make(map[string]interface{})
	md["abc"] = "abc1"
	md["ab"] = "abc"
	err = c.MultiHset("hset", md)
	if err != nil {
		println(i, err)
	} else {
		println("is mhset", i)
	}
	m, err := c.MultiHget("hset", "ab", "test1")
	if err != nil {
		println(i, err)
	} else {
		println(m, "is mhget", i)
	}
	time.Sleep(time.Millisecond)
	//c.Close()
}
