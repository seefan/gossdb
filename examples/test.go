package main

import (
	//	"fmt"
	"github.com/seefan/gossdb"
	"log"
	"reflect"
	"runtime"
	"time"
)

func add(i, j int) int {
	k := (j) + (i)
	t := reflect.ValueOf(j)
	log.Println(t.CanSet())
	return k
}
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
	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	pool.Encoding = true
	client, err := pool.NewClient()
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = client.Set("keys", time.Now(), 1)
	log.Println(err)
	//time.Sleep(time.Second * 3)
	v, err := client.Get("keys")
	log.Println(err, v)
	var test time.Time
	log.Println(err)
	err = v.As(&test)
	log.Println(err, test, v)
	err = client.Hset("set", "key", 132)
	log.Println(err)
	//client.Client.Close()
	v, err = client.Hget("set", "key")
	log.Println(v, err)
	bv, err := client.Hexists("set", "key")
	log.Println(bv, err)
	err = client.Hdel("set", "key")
	log.Println(err)
	err = client.Hclear("set")
	log.Println(err)
	client.Qclear("queue")
	size, err := client.Qpush("queue", 1, 2, 3, test)
	log.Println(err, size)
	size, err = client.Qpush("queue", 3, 2, 1)
	log.Println(err, size)
	v, err = client.Qpop_front("queue")
	log.Println(err, v)
	v, err = client.Qpop_back("queue")
	log.Println(err, v)
	vs, err := client.Qrange("queue", 0, 6)
	log.Println(err, vs)
	vs, err = client.Qslice("queue", 0, 2)
	log.Println(err, vs)
	size, err = client.Qtrim("queue", 1)
	log.Println(err, size)
	vs, err = client.Qslice("queue", 0, 2)
	log.Println(err, vs)
	//	for i := 0; i < 100; i++ {
	//		go func(idx int) {
	//			log.Println(i, pool.Info())
	//			c, err := pool.NewClient()
	//			if err != nil {
	//				log.Println(idx, err.Error())
	//				return
	//			}
	//			defer c.Close()
	//			c.Set(fmt.Sprintf("test%d", idx), fmt.Sprintf("test%d", idx))
	//			re, err := c.Get(fmt.Sprintf("test%d", idx))
	//			if err != nil {
	//				log.Println(err)
	//			} else {
	//				log.Println(re, "is closed")
	//			}
	//		}(i)
	//		time.Sleep(time.Microsecond)
	//	}
	//	time.Sleep(time.Second * 1)
	pool.Close()
	time.Sleep(time.Second * 1)
}
