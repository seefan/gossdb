package ssdb

import (
	"errors"
	"fmt"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

var (
	//连接池实例
	pool *gossdb.Connectors
)

//启动服务
//
//  config 配置文件名，默认为config.ini
//  返回 error，正常启动返回nil
func Start(cfgs ...*conf.Config) error {
	var cfg *conf.Config
	if len(cfgs) == 0 {
		cfg = &conf.Config{
			Host: conf.Host,
			Port: conf.Port,
		}
	} else {
		cfg = cfgs[0]
	}
	conn, err := gossdb.NewPool(cfg)
	if err != nil {
		return err
	}
	pool = conn
	return nil
}

//关闭服务
//
//  返回 error，正常启动返回nil
func Close() {
	if pool != nil {
		pool.Close()
	}
	pool = nil
}

//获取一个连接
//
//  返回 *gossdb.Client
//  返回 error，如果获取到连接就返回nil
func Client() (*gossdb.Client, error) {
	if pool == nil {
		return nil, errors.New("SSDB not initialized")
	}
	return pool.NewClient()
}

//连接的简单使用方法
//
// fn func(c *gossdb.Client) error 实际业务的函数，输入参数为client，输出为error
// 返回 error 可能的错误
//
//    示例：
//
//    ssdb.Simple(func(c *gossdb.Client) (err error) {
//    	err=c.Set("test", "hello world")
//    	err=c.Get("test")
//    	return
//    })
func Simple(fn func(c *gossdb.Client) error) error {
	if client, err := Client(); err == nil {
		if err = fn(client); err != nil {
			if e := client.Close(); e != nil {
				return fmt.Errorf("simple client close error,cause is %s", e.Error())
			}
			return err
		}
		return client.Close()
	} else {
		return err
	}
}
