package main

import (
	"fmt"
	"github.com/seefan/gossdb"
	//	"github.com/seefan/gossdb/ssdb"
	"log"
	"runtime"
	"sync"
	"time"
)

func main() {
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

	ip := "192.168.56.101"
	port := 8888

	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:        ip,
		Port:        port,
		MaxPoolSize: 20,
		MinPoolSize: 20,
		MaxWaitSize: 20000,
	})

	if err != nil {
		fmt.Errorf("error new pool %v", err)
		return
	}
	gossdb.Encoding = true

	client, err := pool.NewClient()
	if err != nil {
		fmt.Errorf("error new pool %v", err)
		return
	}
	defer client.Close()
	//v, err := client.Qpush("test")
	//client.Set("a", "heldfsdlo1")
	//client.Set("b", "hello2")
	//client.Set("keys", "hello")
	//v, err := client.Get("a")
	//log.Println(v, err)
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
	//v, err = client.Hget("set", "key")
	//log.Println(v, err)
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
	//	log.Println("----------------")
	now := time.Now()
	wait := new(sync.WaitGroup)
	sk := new(Success)

	run(wait, pool, sk)
	//log.Println("get client", pool.Info())

	wait.Wait()
	println("time is ", time.Since(now).String(), sk.Show())
	//	log.Println(pool.Info())
	//	time.Sleep(time.Second * 10)
	//	pool.Close() //连接可能未处理完
	//	log.Println(pool.Info())
	//	time.Sleep(time.Second * 10)
}

type Success struct {
	count   int
	success int
	lock    sync.Mutex
}

func (s *Success) Add() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.count += 1
}
func (s *Success) Ok() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.success += 1
}
func (s *Success) Show() string {
	return fmt.Sprintf("count:%d,success:%d", s.count, s.success)
}

func run(wait *sync.WaitGroup, pool *gossdb.Connectors, su *Success) {

	for i := 0; i < 14000; i++ {
		wait.Add(1)
		go func(idx int) {
			defer wait.Done()
			su.Add()
			c, err := pool.NewClient()
			if err != nil {
				//log.Println(err.Error(), idx, pool.Info())
				return
			}
			defer c.Close()
			//err = c.Set(fmt.Sprintf("test%d", idx), fmt.Sprintf("test%d", idx))
			//if err != nil {
			//	log.Println(err)
			//}
			_, err = c.Get(fmt.Sprintf("test%d", idx))
			if err != nil {
				log.Println(err, "get client")
			} else {
				//log.Println(idx, re, "close client", pool.Info())
			}
			//time.Sleep(time.Millisecond * 5)
			su.Ok()
			//
		}(i)
	}
}
