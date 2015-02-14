### gossdb

***

自已用ssdb时开发的，共享出来。

功能列表
* 继承官方连接方式。
* 支持连接池。
* 支持 Set 和 Get 方法，支持 Do 方法。
* 已支持 Hset 和 Hget 方法

更多说明请见[这里](https://gowalker.org/github.com/seefan/gossdb)

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
