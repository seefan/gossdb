package ssdb

import (
	"fmt"
	"github.com/Unknwon/goconfig"
	log "github.com/cihub/seelog"
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

var (
	//连接池实例
	pool *gossdb.Connectors
)

func getConfig(c *goconfig.ConfigFile, name string) conf.Config {
	cfg := conf.Config{
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
	}
	return cfg
}

//启动服务
//
//  config 配置文件名，默认为config.ini
//  返回 error，正常启动返回nil
func Start(config ...string) error {
	log.Info("SSDB连接池启动")
	configName := "config.ini"
	if len(config) > 0 {
		configName = config[0]
	}
	var cfg conf.Config
	cf, err := goconfig.LoadConfigFile(configName)
	if err != nil {
		log.Warnf("未找到SSDB的配置文件%s，将使用默认值启动", configName)
		cf = new(goconfig.ConfigFile)
	}
	cfg = getConfig(cf, "ssdb")
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
	log.Info("SSDB连接池已经结束")
}

//获取一个连接
//
//  返回 *gossdb.Client
//  返回 error，如果获取到连接就返回nil
func Client() (*gossdb.Client, error) {
	if pool == nil {
		return nil, fmt.Errorf("SSDB连接池还未初始化")
	}
	return pool.NewClient()
}
