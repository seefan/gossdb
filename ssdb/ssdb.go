package ssdb

import (
	"fmt"

	"github.com/Unknwon/goconfig"
	log "github.com/cihub/seelog"
	"github.com/seefan/gossdb"
)

var (
	//连接池实例
	pool *gossdb.Connectors
)

//启动服务
//
//  config 配置文件名，默认为config.ini
//  返回 error，正常启动返回nil
func Start(config ...string) error {
	log.Info("SSDB连接池启动")
	if len(config) > 0 {
		ConfigName = config[0]
	}
	cfg, err := goconfig.LoadConfigFile(ConfigName)
	if err == nil {
		Host = cfg.MustValue("ssdb", "host", Host)
		Port = cfg.MustInt("ssdb", "port", Port)
		GetClientTimeout = cfg.MustInt("ssdb", "getclienttimeout", GetClientTimeout)
		MaxPoolSize = cfg.MustInt("ssdb", "maxpoolsize", MaxPoolSize)
		MinPoolSize = cfg.MustInt("ssdb", "minpoolsize", MinPoolSize)
		AcquireIncrement = cfg.MustInt("ssdb", "acquireincrement", AcquireIncrement)
		MaxIdleTime = cfg.MustInt("ssdb", "maxidletime", MaxIdleTime)
		MaxWaitSize = cfg.MustInt("ssdb", "maxwaitsize", MaxWaitSize)
		HealthSecond = cfg.MustInt("ssdb", "healthsecond", HealthSecond)
	} else {
		log.Warnf("未找到SSDB的配置文件%s，将使用默认值启动", ConfigName)
	}
	conn, err := gossdb.NewPool(&gossdb.Config{
		//ssdb的ip或主机名
		Host:             Host,
		Port:             Port,
		GetClientTimeout: GetClientTimeout,
		MaxPoolSize:      MaxPoolSize,
		MinPoolSize:      MinPoolSize,
		AcquireIncrement: AcquireIncrement,
		MaxWaitSize:      MaxWaitSize,
		HealthSecond:     HealthSecond,
	})
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
