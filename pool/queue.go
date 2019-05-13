package pool

import "sync"

//queue for available index
//可用连接的队列
type Queue struct {
	pos    int //available pos
	putPos int //exchange pos
	value  []int
	size   int
	//lock
	lock sync.Mutex
}

//建立一个队列
func newQueue(size int) *Queue {
	v := make([]int, size)
	for i := 0; i < size; i++ {
		v[i] = i
	}
	return &Queue{
		value: v,
		pos:   size - 1,
		size:  size,
	}
}

//check available index
func (q *Queue) Exists() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pos >= 0
}

//Pop get a index
//
//  @return int index
//
//获取一个可以连接的位置
func (q *Queue) Pop() (re int) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.pos < 0 {
		return -1
	}

	re = q.value[q.pos]
	q.value[q.pos] = -1
	q.pos--
	return
}

//Put return a index
//
//  @param i index
//  @return int pos
//
//归还索引值
func (q *Queue) Put(i int) int {
	q.lock.Lock()
	defer q.lock.Unlock()
	pos := q.pos + 1
	if q.putPos < int(pos) {
		q.value[pos] = q.value[q.putPos]
		q.value[q.putPos] = i
		q.putPos++
	} else {
		q.value[pos] = i
		q.putPos = 0
	}
	q.pos = pos
	return q.pos
}
