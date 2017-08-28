package main

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	//	"github.com/seefan/gossdb/ssdb"
	"testing"
)

func Test_set(t *testing.T) {
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "192.168.56.101",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	c, err := pool.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Set("test", "hello world.")
	re, err := c.Get("test")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(re, "is get")
	}
	//设置10 秒过期
	c.Set("test1", 1225, 10)
	//取出数据，并指定类型为 int
	re, err = c.Get("test1")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(re.Int(), "is get")
	}
	m, err := c.MultiGet("test", "test1")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(m, "is mget")
	}
}
func Test_hset(t *testing.T) {
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "192.168.56.101",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	c, err := pool.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Hset("hset", "test", "hello world.")
	re, err := c.Get("test")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(re, "is get")
	}
	//设置10 秒过期
	c.Hset("hset", "test1", 1225)
	//取出数据，并指定类型为 int
	re, err = c.Hget("hset", "test1")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(re.Int(), "is hget")
	}
	md := make(map[string]interface{})
	md["abc"] = "abc1"
	md["ab"] = "abc"
	err = c.MultiHset("hset", md)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("is mhset")
	}
	m, err := c.MultiHget("hset", "ab", "test1")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(m, "is mhget")
	}
}

func Test_zset(t *testing.T) {
	pool, err := gossdb.NewPool(&conf.Config{
		Host:             "192.168.56.101",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	c, err := pool.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Zset("zset", "test", 21)
	re, err := c.Zget("zset", "test")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(re, "is get")
	}
	//设置10 秒过期
	c.Zset("zset", "test1", 1225)
	//取出数据，并指定类型为 int
	re, err = c.Zget("zset", "test1")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(re, "is zset")
	}
	md := make(map[string]int64)
	md["abc"] = 123
	md["ab"] = 4121
	err = c.MultiZset("zset", md)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("is zhset")
	}
	m, err := c.MultiZget("zset", "ab", "test1")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(m, "is zset")
	}
}
