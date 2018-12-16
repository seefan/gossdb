package main

import (
	"github.com/seefan/gossdb/ssdb"
	"testing"
)

func Test_qpush(t *testing.T) {
	err := ssdb.Start("./config.ini")
	if err != nil {
		t.Fatal(t)
	}
	c, err := ssdb.Client()
	if err != nil {
		t.Fatal(t)
	}
	defer c.Close()
	c.Qpush_back("test_qpush", "tk")
	k, err := c.Qpop_back("test_qpush")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(k)
	}
	c.Qpush_front("test_qpush", "tk")
	k, err = c.Qpop_front("test_qpush")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(k)
	}
}

func Test_qpush_array(t *testing.T) {
	err := ssdb.Start("./config.ini")
	if err != nil {
		t.Fatal(t)
	}
	c, err := ssdb.Client()
	if err != nil {
		t.Fatal(t)
	}
	defer c.Close()
	arr := []interface{}{"1", "3", "4"}
	c.Qpush_array("test_qpush_arr", arr)
	k, err := c.QpopArray("test_qpush_arr", 2)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(k)
	}
}

func Test_Qfront(t *testing.T) {
	err := ssdb.Start("./config.ini")
	if err != nil {
		t.Fatal(t)
	}
	c, err := ssdb.Client()
	if err != nil {
		t.Fatal(t)
	}
	defer c.Close()
	k, err := c.Qfront("test_qpush_arr")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(k)
	}
}
