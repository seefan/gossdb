package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/seefan/gossdb"
	golog "github.com/seefan/gossdb/examples/log"

	"runtime"
	"sync"
	"time"
)

func main() {
	defer golog.PrintErr()
	golog.InitSeeLog()
	//	//	i := 12
	//	var v gossdb.Value = "123"
	//	log.Println(v.String())
	//	log.Println(v.Int())
	//	log.Println(v.Int64())
	//	log.Println(v.Int32())
	//	log.Println(v.Int8())
	//	log.Println(v.Int16())
	//	log.Println(v.UInt())
	//	log.Println(v.UInt64())
	//	log.Println(v.UInt32())
	//	log.Println(v.UInt8())
	//	log.Println(v.UInt16())
	//	log.Println(v.Bool())
	//	log.Println(v.Float32())
	//	log.Println(v.Time())
	//	log.Println(v.Duration())
	//	log.Println(v.Bytes())
	//	return
	runtime.GOMAXPROCS(runtime.NumCPU())
	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
		GetClientTimeout: 10,
		MaxWaitSize:      1000,
		MaxIdleTime:      1,
		HealthSecond:     2,
	})
	if err != nil {
		log.Debug(err)
		return
	}
	gossdb.Encoding = true
	//	client, err := pool.NewClient()
	//	if err != nil {
	//		log.Println(err.Error())
	//		return
	//	}
	//	client.Set("a", "hello1")
	//	client.Set("b", "hello2")
	//	client.Set("keys", "hello")
	//	v, err := client.Rscan("z", "", 100)
	//	log.Println(v, err)
	//	err = client.Set("keys", "hello")
	//	log.Println(err)
	//	v, err := client.Setbit("keys", 3, 0)
	//	log.Println(v, err)
	//	v, err = client.Getbit("keys", 3)
	//	log.Println(v, err)
	//client.Del("keys")
	//	v, err := client.Setnx("keys", time.Now())
	//	log.Println(v, err)
	//	v, err = client.Get("keys")
	//	log.Println(err, v)
	//	err = client.Set("keys", time.Now(), 1)
	//	log.Println(err)
	//	//time.Sleep(time.Second * 3)
	//	v, err := client.Get("keys")
	//	log.Println(err, v)
	//	var test time.Time
	//	log.Println(err)
	//	err = v.As(&test)
	//	log.Println(err, test, v)
	//	err = client.Hset("set", "key", 132)
	//	log.Println(err)
	//	//client.Client.Close()
	//	v, err = client.Hget("set", "key")
	//	log.Println(v, err)
	//	v, err = client.Getset("keys", "key1")
	//	log.Println(v, err)
	//	v, err = client.Getset("keys", "key2")
	//	log.Println(v, err)
	//	bv, err := client.Hexists("set", "key")
	//	log.Println(bv, err)
	//	i, err := client.Ttl("set")
	//	log.Println("ttl", i, err)
	//	err = client.Hdel("set", "key")
	//	log.Println(err)
	//	err = client.Hclear("set")
	//	log.Println(err)
	//	client.Qclear("queue")
	//	size, err := client.Qpush("queue", 1, 2, 3, test)
	//	log.Println(err, size)
	//	size, err = client.Qpush("queue", 3, 2, 1)
	//	log.Println(err, size)
	//	v, err = client.Qpop_front("queue")
	//	log.Println(err, v)
	//	v, err = client.Qpop_back("queue")
	//	log.Println(err, v)
	//	vs, err := client.Qrange("queue", 0, 6)
	//	log.Println(err, vs)
	//	vs, err = client.Qslice("queue", 0, 2)
	//	log.Println(err, vs)
	//	size, err = client.Qtrim("queue", 1)
	//	log.Println(err, size)
	//	vs, err = client.Qslice("queue", 0, 2)
	//	log.Println(err, vs)
	//	i, err = client.Incr("incr", 1)
	//	log.Println(i, err)
	//	i, err = client.Incr("incr", 4)
	//	log.Println(i, err)
	//	mm := make(map[string]interface{})
	//	mm["a"] = 1
	//	mm["b"] = 11
	//	mm["a1"] = 1
	//	mm["b11"] = 11
	//	mm["a22"] = 1
	//	mm["b22"] = 11
	//	err = client.MultiSet(mm)
	//	log.Println(err)

	//	err = client.MultiDel("a", "b", "a1")
	//	log.Println(err)
	//	vm, err := client.MultiGet("a", "b", "a1")
	//	log.Println(vm, err)
	log.Debugf("----------------")

	for i := 0; i < 100; i++ {
		go func(idx int) {
			//log.Println(idx, "get client", pool.Info())
			c, err := pool.NewClient()
			if err != nil {
				log.Debugf(err.Error(), idx)
				return
			}
			defer c.Close()
			err = c.Set(fmt.Sprintf("test%d", idx), fmt.Sprintf("test%d", idx))
			if err != nil {
				log.Error(err)
			}
			re, err := c.Get(fmt.Sprintf("test%d", idx))
			if err != nil {
				log.Debug(err, re, "close client")
			} else {
				log.Debug(idx, "close client")
			}
		}(i)
		//time.Sleep(time.Millisecond)
	}
	time.Sleep(time.Second * 10)
	log.Debug(pool.Info())
	time.Sleep(time.Second * 10)
	pool.Close() //连接可能未处理完
	log.Debugf(pool.Info())
	time.Sleep(time.Second * 10)
}

func test1() {
	wait := sync.WaitGroup{}
	locker := new(sync.Mutex)
	cond := sync.NewCond(locker)

	for i := 0; i < 3; i++ {
		go func(i int) {
			defer wait.Done()
			wait.Add(1)
			cond.L.Lock()
			fmt.Println("Waiting start...", i)
			cond.Wait()
			fmt.Println("Waiting end...", i)
			cond.L.Unlock()

			fmt.Println("Goroutine run. Number:", i)
		}(i)
	}

	time.Sleep(2e9)
	cond.L.Lock()
	cond.Signal()
	cond.L.Unlock()

	time.Sleep(2e9)
	cond.L.Lock()
	cond.Signal()
	cond.L.Unlock()

	time.Sleep(2e9)
	cond.L.Lock()
	cond.Signal()
	cond.L.Unlock()

	wait.Wait()
}
