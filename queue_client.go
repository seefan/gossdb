/*
@Time : 2019-04-30 20:42
@Author : seefan
@File : queue
@Software: gossdb
*/
package gossdb

import "sync"

type Queue struct {
	pop   int
	put   int
	ava   int
	value []int
	size  int
	//lock
	lock sync.Mutex
}

func newQueue(size int) *Queue {
	return &Queue{
		value: make([]int, size),
		size:  size,
	}
}
func (q *Queue) Pop() (re int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.ava == 0 {
		return -1
	}
	if q.pop >= q.size {
		q.pop = 0
	}
	re = q.value[q.pop]
	q.pop++
	q.ava--
	return
}
func (q *Queue) Put(i int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.ava++
	if q.put >= q.size {
		q.put = 0
	}
	q.value[q.put] = i
	q.put++
}
