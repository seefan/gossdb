# gossdb

[English Readme](https://gowalker.org/github.com/seefan/gossdb)
### 功能列表

* 参考官方驱动开发，增加连接池支持，改进协议实现方式，提高了数据吞吐量
* 支持 set 相关函数
* 支持 zset 相关函数
* 支持 hset 相关函数
* 支持 queue 相关函数
* 支持 multi 相关函数
* 支持返回值类型转换，可以方便的把从ssdb中取到的内容转化为指定类型
* 支持对象json的序列化，只需要开启Encoding选项
* 支持连接自动回收，支持无错误获取连接，代码调用更简便

### 2.0主要改进
* 修改所有函数名字，使其符合golang编码规范，通过 golint 验证
* 改进协议实现方式，提高解析效率
* 改进连接池方式，提高连接池的存取效率。连接池由原来的单一连接池，改为块状池，每个块都是一个独立的连接池，多个连接池协作，减少锁的竞争时间
* 支持连接自动回收，支持无错误获取连接，代码调用更简便。原来使用连接必须判断连接是否获取成功，并手工关闭，现在可以省略掉这部分重复代码，使得您更专注于业务逻辑
* 解决高并发时的内存泄漏问题，主要是通过回收计时器和重用连接来解决

### 性能测试，仅作参考，并不实际操作ssdb，只打开连接然后立即回收
* BenchmarkConnectors_NewClient10-4   	     5000000	       244 ns/op
* BenchmarkConnectors_NewClient100-4     	 5000000	       215 ns/op
* BenchmarkConnectors_NewClient1000-4   	 5000000	       281 ns/op
* BenchmarkConnectors_NewClient5000-4   	 5000000	       282 ns/op

### 主要配置项

* Host string //ssdb的ip或主机名
* Port int //ssdb的端口
* GetClientTimeout int //获取连接超时时间，单位为秒。默认值: 5
* ReadWriteTimeout int //连接读写超时时间，单位为秒。默认值: 60
* WriteTimeout int //连接写超时时间，单位为秒，如果不设置与ReadWriteTimeout会保持一致。默认值: 0
* ReadTimeout int //连接读超时时间，单位为秒，如果不设置与ReadWriteTimeout会保持一致。默认值: 0
* MaxPoolSize int //最大连接个数。默认值: 100，PoolSize的整数倍，不足的话自动补足。
* MinPoolSize int //最小连接个数。默认值: 20，PoolSize的整数倍，不足的话自动补足。
* PoolSize int //连接池块的连接数。默认值: 20，连接池扩展和收缩时，以此值步进，可根据机器性能调整。
* MaxWaitSize int //最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
* HealthSecond int //连接池内缓存的连接状态检查时间隔，单位为秒。默认值: 30
* Password string //连接的密钥
* WriteBufferSize int //连接写缓冲，默认为8k，单位为kb
* ReadBufferSize int //连接读缓冲，默认为8k，单位为kb
* RetryEnabled bool //是否启用重试，设置为true时，如果请求失败会再重试一次。默认值: false
* ConnectTimeout int //创建连接的超时时间，单位为秒。默认值: 5
* AutoClose bool //是否自动回收连接，如果开启后，获取的连接在使用后立即会被回收，所以不要重复使用。
* Encoding bool //是否开启自动序列化

更多说明请见[这里](https://gowalker.org/github.com/seefan/gossdb)

所有的API基于ssdb的原始API用法，只针对go的特点增加部分方法。所以也可以参照官方文档使用。

引入：

    import "github.com/seefan/gossdb/v2"

示例1：使用自动关闭

    //打开连接池，使用默认配置,Host=127.0.0.1,Port=8888,AutoClose=true
	if err := gossdb.Start(); err != nil {
		panic(err)
	}
	//别忘了结束时关闭连接池，当然如果你没有关闭，ssdb也会因错误中断连接的
	defer gossdb.Shutdown()
	//使用连接，因为AutoClose为true，所以我们没有手工关闭连接
	//gossdb.Client()为无错误获取连接方式，所以可以在获取连接后直接调用其它操作函数，如果获取连接出错或是调用函数出错，都会返回err
	//这时里要注意gossdb.Client()后要立即接调用的函数，这个函数执行完后ssdb的连接会被回收，所以不要重复使用。
	if v, err := gossdb.Client().Get("a"); err == nil {
		println(v.String())
	} else {
		println(err.Error())
	}

调用起来是不是简单很多^_^

示例2： 不使用自动关闭，适用于一个连接进行多个请求的方式

    //使用配置文件，没有将AutoClose设置为true
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
    if v, err := c.Get("b"); err == nil {
		println(v.String())
	} else {
		println(err.Error())
	}
