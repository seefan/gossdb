package pool

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/seefan/gossdb/conf"
)

func BenchmarkConnectors_NewClient(b *testing.B) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    10,
		MaxPoolSize: 200,
	})
	err := pool.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c, err := pool.NewClient()
			if err == nil {
				//_, _ = c.Info()
				c.Close()
			} else {
				//b.Error(err)
			}
		}
	})

	pool.Close()
}
func Test1(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     20,
		MinPoolSize:  10,
		MaxPoolSize:  10,
		HealthSecond: 2,
	})
	err := pool.Start()
	if err != nil {
		panic(err)
	}

	c, err := pool.NewClient()
	if err == nil {
		_, _ = c.Get("a")
		c.Close()
	} else {
		t.Error(err)
	}

	pool.Close()
}
func Test1000(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     10,
		MinPoolSize:  20,
		MaxPoolSize:  100,
		HealthSecond: 2,
	})
	err := pool.Start()
	if err != nil {
		panic(err)
	}
	var wait sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				c, err := pool.NewClient()
				if err == nil {
					_, _ = c.Get("a")
					c.Close()
				} else {
					t.Error(err)
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
	time.Sleep(time.Second * 20)
	pool.Close()
}

func BenchmarkConnectors_NewClient100(b *testing.B) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    20,
		MaxPoolSize: 100,
	})
	err := pool.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c, err := pool.NewClient()
			if err == nil {
				//_, _ = c.Info()
				c.Close()
			} else {
				//b.Error(err)
			}
		}
	})

	pool.Close()
}
func BenchmarkConnectors_NewClient1000(b *testing.B) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    20,
		MaxPoolSize: 500,
	})
	err := pool.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c, err := pool.NewClient()
			if err == nil {
				//_, _ = c.Info()
				c.Close()
			} else {
				//b.Error(err)
			}
		}
	})

	pool.Close()
}
func BenchmarkConnectors_NewClient5000(b *testing.B) {
	pool := NewConnectors(&conf.Config{
		Host:        "127.0.0.1",
		Port:        8888,
		MaxWaitSize: 10000,
		PoolSize:    20,
		MaxPoolSize: 1000,
	})
	err := pool.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.SetParallelism(5000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c, err := pool.NewClient()
			if err == nil {
				//_, _ = c.Info()
				c.Close()
			} else {
				//b.Error(err)
			}
		}
	})

	pool.Close()
}
func TestCheck(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     10,
		MinPoolSize:  10,
		MaxPoolSize:  10,
		HealthSecond: 2,
	})
	err := pool.Start()
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	for {
		c, err := pool.NewClient()
		if err == nil {
			v, err := c.Get("a")
			log.Println(v, err)
			c.Close()
		} else {
			t.Error(err)
		}
	}

}
func TestAutoClose1(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     10,
		MinPoolSize:  10,
		MaxPoolSize:  10,
		HealthSecond: 2,
		AutoClose:    true,
	})
	//
	v, err := pool.GetClient().Get("a")
	t.Log(v, err)
}
func TestAutoClose2(t *testing.T) {
	pool := NewConnectors(&conf.Config{
		Host:         "127.0.0.1",
		Port:         8888,
		MaxWaitSize:  10000,
		PoolSize:     10,
		MinPoolSize:  10,
		MaxPoolSize:  10,
		HealthSecond: 2,
		AutoClose:    true,
	})
	//

	err := pool.Start()
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	for i := 0; i < 100; i++ {
		c, err := pool.NewClient()
		if err != nil {
			panic(err)
		}
		v, err := c.Get("a")
		t.Log(v, err)
	}
}
