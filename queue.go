/*
@Time : 2019-04-30 20:42
@Author : seefan
@File : queue
@Software: gossdb
*/
package gossdb

type Queue struct {
	pos    int
	putPos int
	value  []int
	size   int
}

func newQueue(size int) *Queue {
	return &Queue{
		value: make([]int, size),
		pos:   0,
		size:  size,
	}
}

func (q *Queue) Pop() (re int) {
	if q.pos <= 0 {
		return -1
	}
	q.pos--
	re = q.value[q.pos]
	q.value[q.pos] = -1
	return
}
func (q *Queue) Add(i int) {
	if q.pos < q.size {
		q.value[q.pos] = i
		q.pos++
	}
}
func (q *Queue) Put(i int) int {
	if q.putPos < q.pos {
		q.value[q.pos] = q.value[q.putPos]
		q.value[q.putPos] = i
		q.putPos++
	} else {
		q.value[q.pos] = i
		q.putPos = 0
	}
	q.pos++
	return q.pos

}
