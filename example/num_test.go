package main

import (
	//"github.com/seefan/gossdb"
	"testing"
	"time"
)

func Test_ToNum(t *testing.T) {
	byt := []byte{51, 52, 54, '\r', '\n'}
	t.Log(byt)
	t.Log(byt[1:2])

}
func TestTime(t *testing.T) {
	now := time.Now()
	t.Log(now.String())
	num := now.Unix()
	t.Log(num)
	nt := time.Unix(num, 0)
	t.Log(nt.String())
}
