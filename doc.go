//Package gossdb the client, derived from the official client, supports connection pooling and USES the same conventions as most clients.
//
//     refer to official driver development, add connection pool support, improve protocol implementation, and improve data throughput
//     support for set dependent functions
//     supports zset correlation functions
//     supports hset correlation functions
//     supports queue related functions
//     supports multi correlation functions
//     supports return value type conversion, which makes it easy to convert content from SSDB to the specified type
//     supports serialization of json objects, just open the gossdb.encoding option
//     support automatic connection recovery, support error free connection acquisition, code call is easier
//
//  Major improvements in 2.0
//
//     modify the names of all functions to make them conform to the golang coding program and pass the golint verification
//     improve protocol implementation to improve parsing efficiency
//     improve connection pool mode to improve access efficiency of connection pool. Instead of a single connection pool, the connection pool is a block pool,
//     each block is a separate connection pool, and multiple connection pools collaborate to reduce lock contention time
//     support automatic connection recovery, support error free connection acquisition, code call is easier. Instead of having to determine if the connection
//     was successful and close it manually, you can omit this duplicated code and focus on the business logic
//     the memory leak problem with high concurrency is solved primarily by recycling timers and reusing connections
//
//  The main configuration item
//
//     // SSDB IP or host name
//     the Host string
//     // SSDB port
//     the Port int
//     // gets the connection timeout in seconds. Default: 5
//     GetClientTimeout int
//     // connection read/write timeout in seconds. Default: 60
//     ReadWriteTimeout int
//     // the connection write timeout, in seconds, is the same as the ReadWriteTimeout if not set. Default: 0
//     WriteTimeout int
//     // the connection read timeout, in seconds, is the same as the ReadWriteTimeout if not set. Default: 0
//     ReadTimeout int
//     // maximum number of connections. Default value: 100, integer multiple of PoolSize, if not enough, it will be filled automatically.
//     MaxPoolSize int
//     // minimum number of connections. Default value: 20, integer multiple of PoolSize.
//     MinPoolSize int
//     // the number of connections to the pool block. Default value: 20, when connection pool is expanded and contracted, step by this value, which can be adjusted according to machine performance.
//     PoolSize int
//     // the maximum number of waits. When the connection pool is full, the new connection will wait for the connection in the pool to be released before //    ontinuing. This value limits the maximum number of waits. Default: 1000
//     MaxWaitSize int
//     // the connection status check interval for the cache in the connection pool is in seconds. Default: 30
//     HealthSecond int
//     // key for connection
//     the Password string
//     // connection write buffer, default 8k, in KB
//     WriteBufferSize int
//     // connection read buffer, default 8k, in KB
//     ReadBufferSize int
//     // if retry is enabled, set to true and try again if the request fails. Default: false
//     RetryEnabled bool
//     // the timeout for creating a connection in seconds. Default: 5
//     ConnectTimeout int
//     // auto close
//     AutoClose bool
//
//  More instructions please see [here] (https://gowalker.org/github.com/seefan/gossdb)
//
//  All apis are essentially faithful to the original API usage of SSDB, with only a few methods added for go features. So you can also refer to the official //  documentation.
//
//  Example 1: use automatic shutdown
//
//      //open the connection pool, using the default configuration,Host=127.0.0.1,Port=8888,AutoClose=true
//  	if err := gossdb.Start(); err != nil {
//  		panic(err)
//  	}
//  	//don't forget to close the connection pool at the end of the session. Of course, if you don't close the connection, the SSDB will also break the //  connection due to an error
//  	defer gossdb.Shutdown()
//  	//use the connection, since AutoClose is true, we did not close the connection manually
//  	//gossdb.client () is error-free connection mode, so it can directly call other operation functions after obtaining the connection. If the connection //  is wrong or the calling function is wrong, it will return err
//  	if v, err := gossdb.Client().Get("a"); err == nil {
//  		println(v.String())
//  	} else {
//  		println(err.Error())
//  	}
//
//  Call up is many simple  ^_^
//
//  Example 2: does not use automatic shutdown, works for a way to connect multiple requests at once
//
//  	// with the configuration file, AutoClose is not set to true
//      err := gossdb.Start(&conf.Config{
//  		Host: "127.0.0.1",
//  		Port: 8888,
//  	})
//  	if err != nil {
//  		panic(err)
//  	}
//  	defer gossdb.Shutdown()
//  	c, err := gossdb.NewClient()
//  	if err != nil {
//  		panic(err)
//  	}
//  	defer c.Close()
//  	if v, err := c.Get("a"); err == nil {
//  		println(v.String())
//  	} else {
//  		println(err.Error())
//  	}
//      if v, err := c.Get("b"); err == nil {
//  		println(v.String())
//  	} else {
//  		println(err.Error())
//  	}
package gossdb
