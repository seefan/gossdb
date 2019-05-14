/*
@Time : 2019-05-07 16:58
@Author : seefan
@File : test
@Software: gossdb
*/
package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

func main() {
	p, err := gossdb.NewPool(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    10,
		MinPoolSize: 100,
		MaxPoolSize: 100,
		AutoClose:   true,
	})
	if err != nil {
		panic(err)
	}
	go func() {
		err := http.ListenAndServe(":9999", nil)
		if err != nil {
			panic(err)
		}
	}()
	for i := 0; i < 100; i++ {
		go func() {
			for {
				//failed := 0
				for j := 0; j < 1000; j++ {
					//if _, err := p.GetClient().Get("a"); err != nil {
					//	//println(goerr.Error(err).Trace())
					//	failed++
					//	println(failed, j)
					//}
					c, err := p.NewClient()
					if err != nil {
						println(err.Error(), p.Info())
					} else {
						//if _, err := c.Get("a"); err != nil {
						//	println(goerr.Error(err).Trace())
						//}
						c.Close()
					}
				}
				//println(failed, p.Info())
				//time.Sleep(time.Second)

			}
		}()
	}
	bs := make([]byte, 1)
	os.Stdin.Read(bs)
	p.Close()
}
