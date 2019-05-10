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

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/ssdb"
)

func main() {
	test()
	return
	p, err := gossdb.NewPool(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    10,
		MinPoolSize: 10,
		MaxPoolSize: 50,
		AutoClose:   true,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i++ {
		go func() {
			for {
				failed := 0
				for j := 0; j < 1000; j++ {
					if _, err := p.GetClient().Get("a"); err != nil {
						//println(goerr.Error(err).Trace())
						failed++
						println(failed, j)
					}
					//c, err := p.NewClient()
					//if err != nil {
					//	println(err.Error(), p.Info())
					//} else {
					//	if _, err := c.Get("a"); err != nil {
					//		println(goerr.Error(err).Trace())
					//	}
					//	c.Close()
					//}
				}
				println(failed, p.Info())
				//time.Sleep(time.Minute)

			}
		}()
	}
	bs := make([]byte, 1)
	os.Stdin.Read(bs)
	p.Close()
}
func test() {
	if err := ssdb.Start(); err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if v, err := ssdb.ClientAutoClose().Get("a"); err == nil {
			writer.Write(v.Bytes())
		}
	})
	http.ListenAndServe(":8899", nil)
}
