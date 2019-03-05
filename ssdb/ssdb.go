package ssdb

import (
	"errors"

	"github.com/Unknwon/goconfig"
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

var (
	//连接池实例
	pool *gossdb.Connectors
)

func getConfig(c *goconfig.ConfigFile, name string) *conf.Config {
	cfg := &conf.Config{
		Port:             c.MustInt(name, "port", conf.Port),
		Host:             c.MustValue(name, "host", conf.Host),
		HealthSecond:     c.MustInt(name, "health_second", conf.HealthSecond),
		Weight:           c.MustInt(name, "weight", conf.Weight),
		Password:         c.MustValue(name, "password", ""),
		MaxWaitSize:      c.MustInt(name, "max_wait_size", conf.MaxWaitSize),
		AcquireIncrement: c.MustInt(name, "acquire_increment", conf.AcquireIncrement),
		MinPoolSize:      c.MustInt(name, "min_pool_size", conf.MinPoolSize),
		MaxPoolSize:      c.MustInt(name, "max_pool_size", conf.MaxPoolSize),
		GetClientTimeout: c.MustInt(name, "get_client_timeout", conf.GetClientTimeout),
		IdleTime:         c.MustInt(name, "idle_time", conf.IdleTime),
	}
	return cfg
}

//启动服务
//
//  config 配置文件名，默认为config.ini
//  返回 error，正常启动返回nil
func Start(config ...string) error {
	configName := conf.ConfigName
	if len(config) > 0 {
		configName = config[0]
	}

	cf, err := goconfig.LoadConfigFile(configName)
	if err != nil {
		cf = new(goconfig.ConfigFile)
	}
	cfg := getConfig(cf, "ssdb")
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
		return nil, errors.New("SSDB连接池还未初始化")
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
		defer client.Close()
		if err = fn(client); err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}
