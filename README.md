### gossdb

***


功能列表
* 继承官方连接方式。已支持超过99%的官方命令。
* 支持连接池。
* 已支持 set 相关方法
* 已支持 zset 相关方法
* 已支持 hset 相关方法
* 已支持 queue 相关方法
* 已支持返回值类型转换，可以方便的把从ssdb中取到的内容转化为指定类型。

连接池已支持如下参数
* GetClientTimeout int 获取连接超时时间，单位为秒，默认为 5
* MaxPoolSize int 最大连接池个数，默认为 20
* MinPoolSize int 最小连接池数，默认为 5
* AcquireIncrement int  当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
* MaxIdleTime int 最大空闲时间，指定秒内未使用则连接被丢弃。若为0则永不丢弃。默认值: 0
* MaxWaitSize int //最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
* HealthSecond int 健康检查时间隔，单位为秒。默认值: 300。通过ping方法维护连接的可用性，并定期回收长期不用的连接。

更多说明请见[这里](https://gowalker.org/github.com/seefan/gossdb)

所有的API基本上忠于ssdb的原始API用法，只针对go的特点增加部分方法。所以也可以参照官方文档使用。

示例：


	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             "127.0.0.1",
		Port:             6380,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Fatal(err)
		return
	}


	c, err := pool.NewClient()
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer c.Close()
	c.Set("test","hello world.")
	re, err := c.Get("test")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(re, "is get")
	}
	//设置10 秒过期
	c.Set("test1",1225,10)
	//取出数据，并指定类型为 int
	re, err = c.Get("test1")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(re.Int(), "is get")
	}
