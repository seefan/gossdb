package main

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil) // 启动默认的 http 服务，可以使用自带的路由
	}()
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "192.168.56.101",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      500,
		MaxWaitSize:      10000,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	for i := 0; i < 10000; i++ {
		go func() {
			for {
				Test_hset1(pool)
			}
		}()
	}
	time.Sleep(time.Hour)
}
func Test_hset1(pool *gossdb.Connectors) {

	c, err := pool.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()
	c.Hset("hset", "test", "hello world.")
	re, err := c.Get("test")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(re, "is get")
	}
	//log.Println(pool.Info())
	//md := make(map[string]interface{})
	//md["abc"] = "abc1"
	//md["ab"] = "abc"
	//err = c.MultiHset("hset", md)
	//if err != nil {
	//	log.Println(err)
	//} else {
	//	log.Println("is mhset")
	//}
	//m, err := c.MultiHget("hset", "ab", "test1")
	//if err != nil {
	//	log.Println(err)
	//} else {
	//	log.Println(m, "is mhget")
	//}
}
