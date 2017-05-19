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
	//ssdb的ip或主机名
	Host = "127.0.0.1"
	// ssdb的端口
	Port = 8888
	//获取连接超时时间，单位为秒。默认值: 5
	GetClientTimeout = 5
	//最大连接池个数。默认值: 20
	MaxPoolSize = 100
	//最小连接池数。默认值: 5
	MinPoolSize = 5
	//当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
	AcquireIncrement = 5
	//最大空闲时间，指定秒内未使用则连接被丢弃。若为0则永不丢弃。默认值: 0
	MaxIdleTime = 600
	//最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
	MaxWaitSize = 20
	//健康检查时间隔，单位为秒。默认值: 300
	HealthSecond = 300
	//默认配置文件名
	ConfigName = "config.ini"
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
		MaxIdleTime:      MaxIdleTime,
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
