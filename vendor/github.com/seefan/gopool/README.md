# gopool
简单且高效的通用连接池。
任何struct只需要实现IClient接口即可享受本连接池带来的高效率。

### 性能测试 


    测试环境
    MacBook Pro (13-inch, Mid 2012) 
    cpu 2.5 GHz Intel Core i5
    内存 8 GB 1600 MHz DDR3

    goos: darwin
    goarch: amd64
    pkg: github.com/seefan/gopool
    5000000	       262 ns/op
    5000000	       265 ns/op
    5000000	       267 ns/op
    5000000	       271 ns/op
    5000000	       263 ns/op
    3000000	       454 ns/op
    3000000	       440 ns/op
    3000000	       445 ns/op
    3000000	       438 ns/op
    3000000	       429 ns/op
    PASS