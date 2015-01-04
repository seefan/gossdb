package main

import (
	"fmt"
	"github.com/seefan/gossdb"
	"log"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             "127.0.0.1",
		Port:             6380,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	for i := 0; i < 100; i++ {
		go func(idx int) {
			log.Println(i, pool.Info())
			c, err := pool.NewClient()
			if err != nil {
				log.Println(idx, err.Error())
				return
			}
			defer c.Close()
			c.Set(fmt.Sprintf("test%d", idx), fmt.Sprintf("test%d", idx))
			re, err := c.Get(fmt.Sprintf("test%d", idx))
			if err != nil {
				log.Println(err)
			} else {
				log.Println(re, "is closed")
			}
		}(i)
		time.Sleep(time.Microsecond)
	}
	time.Sleep(time.Second * 1)
	pool.Close()
	time.Sleep(time.Second * 1)
}
