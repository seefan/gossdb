package ssdbclient

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/seefan/gossdb/v2/conf"
)

func TestSSDBClient_ping(t *testing.T) {
	cfg := &conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Do("dbsize"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_getset(t *testing.T) {
	cfg := &conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Do("set", "a", "123"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "a"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_int(t *testing.T) {
	cfg := &conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Do("set", "a", 123); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "a"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_uint(t *testing.T) {
	cfg := &conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Do("set", "a", uint32(123)); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "a"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_multi(t *testing.T) {
	cfg := &conf.Config{
		Host:     "127.0.0.1",
		Port:     8888,
		Encoding: true,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Do("multi_set", "a", "abc", "b", "ddd", "c", "eft"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "a"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "b"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	c.EncodingFunc = func(v interface{}) []byte {
		if bs, err := json.Marshal(v); err == nil {
			return bs
		}
		return nil
	}
	if v, err := c.Do("set", "add", []string{"hello", "world"}); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "add"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSSDBClient_time(t *testing.T) {
	cfg := &conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if v, err := c.Do("set", "atime", now); err == nil {
		t.Log(v, now.Unix())
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "atime"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSSDBClient_byte(t *testing.T) {
	cfg := &conf.Config{
		Host: "127.0.0.1",
		Port: 8888,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	bt := byte(126)
	if v, err := c.Do("set", "ab", bt); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "ab"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_pwd(t *testing.T) {
	cfg := &conf.Config{
		Host:     "127.0.0.1",
		Port:     8888,
		Password: "vdsfsfafapaddssrd#@Ddfasfdsfedssdfsdfsd",
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	bt := byte(126)
	if v, err := c.Do("set", "ab", bt); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "ab"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_getBig(t *testing.T) {
	cfg := &conf.Config{
		Host:            "127.0.0.1",
		Port:            8888,
		ReadBufferSize:  8,
		WriteBufferSize: 8,
		ReadTimeout:     300,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	//var s = make([]byte, 5*1024*1024)
	//for i := 0; i < len(s); i++ {
	//	s[i] = 'a'
	//}

	//if v, err := c.Do("set", "big", s); err == nil {
	//	t.Log(v)
	//} else {
	//	t.Error(err)
	//}
	//for i := 0; i < 1000; i++ {
	if v, err := c.Do("hget", "app:0", "1359003378"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	//}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestSSDBClient_getScan(t *testing.T) {
	cfg := &conf.Config{
		Host:            "127.0.0.1",
		Port:            8888,
		ReadBufferSize:  8,
		WriteBufferSize: 8,
		ReadTimeout:     300,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if v, err := c.Do("hscan", "rank:2019-05-18:0:free", "", "", 10000); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

}
func TestSSDBClient_multiget(t *testing.T) {
	cfg := &conf.Config{
		Host:            "127.0.0.1",
		Port:            8888,
		ReadBufferSize:  8,
		WriteBufferSize: 8,
		ReadTimeout:     300,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if v, err := c.Do("multi_hset", "black:0", "0:0:1000565011", "2", "0:0:1001394200", "3"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("multi_hget", "black:0", "0:0:1000565011", "0:0:1001394200"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}

}
func TestSSDBClient_nil(t *testing.T) {
	cfg := &conf.Config{
		Host:            "127.0.0.1",
		Port:            8888,
		ReadBufferSize:  8,
		WriteBufferSize: 8,
		ReadTimeout:     300,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if v, err := c.Do("set", "test_nil", nil); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
	if v, err := c.Do("get", "test_nil"); err == nil {
		t.Log(v)
	} else {
		t.Error(err)
	}
}

func BenchmarkConnectors_conv(b *testing.B) {
	ns := []string{"1283221", "3132", "32331", "92847", "9863232", "93712"}
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := range ns {
				strconv.Atoi(ns[i])
			}
		}
	})
}
