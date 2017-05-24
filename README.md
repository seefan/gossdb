# gossdb


### 功能列表

* 继承官方连接方式。
* 支持连接池。新版本使用了新的连接池，性能比原连接池大约提升10%，并使用了更激进的连接回收方式，防止占用大量连接。
* 已支持 set 相关方法
* 已支持 zset 相关方法
* 已支持 hset 相关方法
* 已支持 queue 相关方法
* 已支持返回值类型转换，可以方便的把从ssdb中取到的内容转化为指定类型。

### 连接池参数

* MaxPoolSize int 最大连接池个数，默认为 20
* MinPoolSize int 最小连接池数，默认为 5
* GetClientTimeout int 获取连接超时时间，单位为秒，默认为 5
* AcquireIncrement int  当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
* MaxWaitSize int //最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000，建议在内存允许的情况下，设置为最大并发数量。
* HealthSecond int 健康检查时间隔，单位为秒。默认值: 5。定期回收长期不用的连接。

### 示例配置

    [ssdb]
    #ssdb的主机IP
    host=127.0.0.1
    #ssdb的端口
    port=8888
    #连接池检查时间间隔
    health_second=5
    #连接密码，默认为空
    password=
    #最大等待数目，当连接池满后，新建连接将等待池中连接释放后才可以继续，本值限制最大等待的数量，超过本值后将抛出异常。默认值: 1000
    max_wait_size=1000
    #当连接池中的连接耗尽的时候一次同时获取的连接数。默认值: 5
    acquire_increment=5
    #最小连接池数。默认值: 5
    min_pool_size=5
    #最大连接池个数。默认值: 20
    max_pool_size=20
    #获取连接超时时间，单位为秒。默认值: 5
    get_client_timeout=5

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

### 更方便的方法
使用ssdb目录下的工具类

    if err := ssdb.Start(); err != nil {
		println("无法连接到ssdb")
		os.Exit(1)
	}
	defer ssdb.Close()
	//获取client
	client, err := ssdb.Client()
	if err != nil {
		println("无法获取连接")
		os.Exit(1)
	}
	defer client.Close()
	client.Set("a", 1)
	client.Get("a")