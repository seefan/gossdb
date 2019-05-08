package ssdb

import (
	"errors"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

var (
	//连接池实例
	pc *gossdb.Connectors
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
	pc = conn
	return nil
}

//关闭服务
//
//  返回 error，正常启动返回nil
func Close() {
	if pc != nil {
		pc.Close()
	}
	pc = nil
}

//获取一个连接
//
//  返回 *gossdb.Client
//  返回 error，如果获取到连接就返回nil
func Client() (*gossdb.Client, error) {
	if pc == nil {
		return nil, errors.New("SSDB not initialized")
	}
	return pc.NewClient()
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
	if pc == nil {
		return errors.New("SSDB not initialized")
	}
	client, err := Client()
	if err != nil {
		return err
	}
	defer client.Close()

	if err = fn(client); err != nil {
		return err
	}
	return nil
}
