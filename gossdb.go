package gossdb

import (
	"errors"

	"github.com/seefan/gossdb/v2/conf"
	"github.com/seefan/gossdb/v2/pool"
)

var (
	//global instance
	//连接池实例
	pooled *pool.Connectors
)

// NewPool start a gossdb pool with the initial parameters.
//
//  @param conf initial parameters
//  @return gossdb pool
//  @return error that may occur on startup. Return nil if successful startup
//
// 通常使用NewPool启动连接池，初始化参数由外部传入。如果启动成功返回一个连接池和nil，如果失败返回nil和发生的错误。
//
func NewPool(conf *conf.Config) (*pool.Connectors, error) {
	//默认值处理
	c := pool.NewConnectors(conf)
	if err := c.Start(); err != nil {
		return nil, err
	}
	return c, nil
}

// Start start a global connection pool. This function can only be started once in the program's lifetime.
// The Shutdown() is called to close the connection pool when the program ends.
//
//  @param cfg initial parameters,optional. Use with default parameters if not specified
//  @return error that may occur on startup. Return nil if successful startup
//
// 启动一个全局的连接池，此函数在程序的生命周期内只能启动一次，程序结束时调用Shutdown函数关闭连接池。
// 参数cfg是可选的，如果没有指定，会使用默认参数连接ssdb，host为127.0.0.1，port为8888,autoclose为true。一般在学习gossdb时用于本地练习。
//
func Start(cfg ...*conf.Config) (err error) {
	var c *conf.Config
	if len(cfg) == 0 {
		c = &conf.Config{
			Host:      "127.0.0.1",
			Port:      8888,
			AutoClose: true,
		}
	} else {
		c = cfg[0]
	}
	pooled, err = NewPool(c)
	return
}

//Shutdown shutdown gossdb pool
//
//关闭连接池。如果你用Start函数启动了连接池，就用这个函数关闭它。
//
func Shutdown() {
	if pooled != nil {
		pooled.Close()
	}
}

//Client returns a connection If no connection is available in the connection pool,
// gossdb creates a temporary connection and returns it, and any action on the connection returns an error.
// This error is used to indicate that the connection pool cannot provide a new connection.
//
//  @return PoolClient
//
//这个函数返回一个连接，并且总是成功返回。如果连接池没有可用的连接时，gossdb会创建一个临时的连接并返回，
//这个连接进行任何操作都会返回一个错误。这个错误用来标记连接池无法提供新连接。
//
func Client() *pool.Client {
	if pooled != nil {
		return pooled.GetClient()
	}
	return nil //故意返回nil，让程序崩溃
}

//NewClient returns a cached connection and possible errors
//
//  @return *PoolClient
//  @return error
//
//这个函数返回一个缓存的连接和一个可能的错误，如果成功返回的错误就为nil。
//
func NewClient() (*pool.Client, error) {
	if pooled == nil {
		return nil, errors.New("gossdb not initialized")
	}
	return pooled.NewClient()
}
