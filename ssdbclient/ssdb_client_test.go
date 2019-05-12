package ssdbclient

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/seefan/gossdb/conf"
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
func TestSSDBClient_stringSlice(t *testing.T) {
	cfg := &conf.Config{
		Host:     "127.0.0.1",
		Port:     8888,
		Encoding: true,
	}
	c := NewSSDBClient(cfg.Default())
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	if v, err := c.Do("multi_set", []string{"a", "abc", "b", "ddd", "c", "eft"}); err == nil {
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
