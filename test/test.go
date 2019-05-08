/*
@Time : 2019-05-07 16:58
@Author : seefan
@File : test
@Software: gossdb
*/
package main

import (
	"net/http"
	"os"
	"runtime/pprof"

	//_ "net/http/pprof"
	"sync"
	"time"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/pool"
)

func main() {

	pool, err := gossdb.NewPool(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 1000,
		PoolSize:    10,
		MinPoolSize: 100,
		MaxPoolSize: 100,
	})
	if err != nil {
		panic(err)
	}
	//远程获取pprof数据

	defer pool.Close()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		c, err := pool.NewClient()
		if err != nil {
			println(err.Error())
			return
		}
		defer c.Close()
		//if rsp, err := c.Get("a"); err == nil {
		//	writer.Write(rsp.Bytes())
		//} else {
		//	println(err)
		//}
		writer.Write([]byte("1"))
	})
	http.ListenAndServe(":8899", nil)
}
func main1() {
	f, e := os.Create("prof.dat")
	if e != nil {
		panic(e)
	}
	defer f.Close()
	_ = pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	p := pool.NewConnectors(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     10,
		MinPoolSize:  20,
		MaxPoolSize:  100,
		HealthSecond: 2,
	})
	err := p.Start()
	if err != nil {
		panic(err)
	}
	var wait sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				c, err := p.NewClient()
				if err == nil {
					//_, _ = c.Get("a")
					c.Close()
				} else {
					println(err.Error())
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
	time.Sleep(time.Second * 20)
	p.Close()
}
