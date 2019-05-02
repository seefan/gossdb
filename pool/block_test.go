/*
@Time : 2019-05-02 16:47
@Author : seefan
@File : block_test.go
@Software: gossdb
*/
package pool

import (
	"testing"
	"time"

	"github.com/seefan/gossdb/ssdb_client"
)

func TestBlock(t *testing.T) {
	var cs []*PooledClient
	for i := 0; i < 5; i++ {
		c := &PooledClient{
			Client: &ssdb_client.SSDBClient{
				Host:             "127.0.0.1",
				Port:             8888,
				ReadWriteTimeout: 30,
				ReadBufferSize:   1024,
				WriteBufferSize:  1024,
			},
		}
		if err := c.Client.Start(); err != nil {
			t.Error(err)
		}
		cs = append(cs, c)
	}
	b := newBlock(cs)
	for i := 0; i < 10; i++ {
		go func() {
			for {
				if c, e := b.Get(); e != nil {
					t.Error(e)
				} else {
					h := c.Client.Ping()
					t.Log(h)
					b.Set(c)
				}
				//time.Sleep(time.Millisecond)
			}
		}()
	}
	time.Sleep(2 * time.Second)
}
func BenchmarkRun(b *testing.B) {
	var cs []*PooledClient
	for i := 0; i < 5; i++ {
		c := &PooledClient{
			Client: &ssdb_client.SSDBClient{
				Host:             "127.0.0.1",
				Port:             8888,
				ReadWriteTimeout: 30,
				ReadBufferSize:   1024,
				WriteBufferSize:  1024,
			},
		}
		if err := c.Client.Start(); err != nil {
			b.Error(err)
		}
		cs = append(cs, c)
	}
	bs := newBlock(cs)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if c, e := bs.Get(); e != nil {
				b.Error(e)
			} else {
				c.Client.Ping()
				//b.Log(h)
				bs.Set(c)
			}
		}
	})
}
