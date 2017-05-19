//从官方客户端派生出来的客户端，支持连接池，使用习惯与大多数客户端保持一致。
//
//  继承官方连接方式。
//  支持连接池。包括连接过期，大小限制，超时等基本功能。
//  支持 json 自动编码
//  支持返回值类型转换
//
//示例：
//
//	pool, err := gossdb.NewPool(&gossdb.Config{
//		Host:             "127.0.0.1",
//		Port:             6380,
//		MinPoolSize:      5,
//		MaxPoolSize:      50,
//		AcquireIncrement: 5,
//	})
//	if err != nil {
//		log.Fatal(err)
//		return
//	}
//
//
//	c, err := pool.NewClient()
//	if err != nil {
//		log.Println(idx, err.Error())
//		return
//	}
//	defer c.Close()
//	c.Set("test","hello world.")
//	re, err := c.Get("test")
//	if err != nil {
//		log.Println(err)
//	} else {
//		log.Println(re, "is get")
//	}
package gossdb
