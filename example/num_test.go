package main

import (
	"github.com/seefan/gossdb"
	"testing"
)

func Test_ToNum(t *testing.T) {
	byt := []byte{51, 52, 54, '\r', '\n'}
	t.Log(byt)
	t.Log(gossdb.ToNum(byt))
}
