package main

import (
	"github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/ssdb"
	"testing"
)

func Test_getset(t *testing.T) {
	err := ssdb.Start(&conf.Config{Host: "127.0.0.1", Port: 8888})
	if err != nil {
		t.Fatal(t)
	}
	c, err := ssdb.Client()
	if err != nil {
		t.Fatal(t)
	}
	defer c.Close()
	c.Set("test_getset", "tk")
	k, err := c.Get("test_getset")
	if err != nil {
		t.Log(k.String() == "tk")
	}
}
