/*
@Time : 2019-04-30 20:42
@Author : seefan
@File : queue
@Software: gossdb
*/
package pool

import (
	"testing"
)

func TestQueue_Pop(t *testing.T) {
	q := newQueue(3)
	for i := 0; i < 3; i++ {
		q.Put(i + 10)
	}
	for i := 0; i < 10; i++ {
		k := q.Pop()
		t.Log(k)
	}
}
