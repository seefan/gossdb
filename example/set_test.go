package main

import (
	"testing"

	"github.com/seefan/gossdb/ssdb"
)

func Test_getset(t *testing.T) {
	err := ssdb.Start()
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
