//Package pool available index queue
package queue

//Queue queue for available index
//可用连接的队列
type Queue struct {
	pos   int //available pos
	value []int
	size  int
}

//建立一个队列
func NewQueue(size int) *Queue {
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

//Available queue available size
func (q *Queue) Available() int {
	return q.size - q.pos
}

//IsEmpty check available index
//
//  @return bool
func (q *Queue) IsEmpty() (re bool) {
	re = q.pos < 0
	return
}

//Pop get a index
//
//  @return int index
//
//获取一个可以连接的位置
func (q *Queue) Pop() (re int) {
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
func (q *Queue) Put(v int) int {
	pos := q.pos + 1
	if pos < q.size {
		q.value[pos] = v
		q.pos = pos
		return pos
	}
	return -1
}
