package main

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"testing"
	"github.com/seefan/gossdb/ssdb"
)

func Test_1(t *testing.T) {
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
}
func Test_2(t *testing.T) {
	err := ssdb.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer ssdb.Close()
	c, err := ssdb.Client()
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
}
