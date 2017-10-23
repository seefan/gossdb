package main

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	//"time"

	log "github.com/cihub/seelog"
	//"net/http"
	_ "net/http/pprof"
)

func main() {
	if logger, err := log.LoggerFromConfigAsFile("./log.xml"); err == nil {
		log.ReplaceLogger(logger)
	}
	defer log.Flush()
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      1,
		MaxPoolSize:      500,
		MaxWaitSize:      10000,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Critical(err)
	}
	defer pool.Close()
	Testhset3(pool)
}

func Testhset3(pool *gossdb.Connectors) {

	c, err := pool.NewClient()
	if err != nil {
		log.Info("create", 0, err)
		return
	}
	defer c.Close()
	c.Set("test", "hello world.")
	if err != nil {
		log.Info(1, err)
	} else {
		log.Info("is hset", 1)
	}
	re, err := c.Get("test")
	if err != nil {
		log.Info(1, err)
	} else {
		log.Info(re, "is get", 1)
	}
	log.Info(pool.Info())

}
