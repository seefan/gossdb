/*
@Time : 2019-05-07 16:58
@Author : seefan
@File : test
@Software: gossdb
*/
package main

import (
	"net/http"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

func main() {
	pool, err := gossdb.NewPool(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 1000,
		PoolSize:    50,
		PoolNumber:  20,
	})
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		c, err := pool.NewClient()
		if err != nil {
			println(err)
			return
		}
		defer c.Close()
		if rsp, err := c.Get("a"); err == nil {
			writer.Write(rsp.Bytes())
		} else {
			println(err)
		}

	})
	http.ListenAndServe(":8899", nil)
}
