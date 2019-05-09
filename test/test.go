/*
@Time : 2019-05-07 16:58
@Author : seefan
@File : test
@Software: gossdb
*/
package main

import (
	"os"
	"time"

	"github.com/seefan/goerr"
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

func main() {
	p, err := gossdb.NewPool(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     10,
		MinPoolSize:  20,
		MaxPoolSize:  100,
		HealthSecond: 2,
		AutoClose:    true,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 1; i++ {
		go func() {
			for {
				for j := 0; j < 1000; j++ {
					if _, err = p.GetClient().Get("a"); err != nil {
						println(goerr.Error(err).Trace())
					}
				}
				time.Sleep(time.Minute)
				println("sleep one minute")
			}
		}()
	}
	bs := make([]byte, 1)
	os.Stdin.Read(bs)
	p.Close()
}
