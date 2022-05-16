package gossdb_test

import (
	"github.com/seefan/gossdb/v2"
	"github.com/seefan/gossdb/v2/conf"
)

//Client
//注意默认配置里连接自动关闭为true，所以此处没有手工关闭连接。
//本地ssdb连接示例。假设host为127.0.0.1，端口为8888
//
func ExampleClient_autoClose() {
	if err := gossdb.Start(); err != nil {
		panic(err)
	}
	defer gossdb.Shutdown()
	if v, err := gossdb.Client().Get("a"); err == nil {
		println(v.String())
	} else {
		println(err.Error())
	}
}

//NewClient
//注意默认配置里没有把连接自动关闭为true，所以此处手工关闭连接。
//本地ssdb连接示例。假设host为127.0.0.1，端口为8888
//
func ExampleNewClient_notAutoClose() {
	err := gossdb.Start(&conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	})
	if err != nil {
		panic(err)
	}
	defer gossdb.Shutdown()
	c, err := gossdb.NewClient()
	if err != nil {
		panic(err)
	}
	defer c.Close()
	if v, err := c.Get("a"); err == nil {
		println(v.String())
	} else {
		println(err.Error())
	}
}
