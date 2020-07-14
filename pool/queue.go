//Package pool available index queue
package pool

//Queue queue for available index
//可用连接的队列
type Queue struct {
	pos    int //available pos
	putPos int //exchange pos
	value  []int
	size   int
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

//Available queue available size
func (q *Queue) Available() int {
	return q.size - q.pos
}

//Empty check available index
//
//  @return bool
func (q *Queue) Empty() (re bool) {
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
	//println("pop", q.pos)
	return
}

//Put return a index
//
//  @param i index
//  @return int pos
//
//归还索引值
func (q *Queue) Put(i int) {
	pos := q.pos + 1
	if q.putPos < pos {
		q.value[pos] = q.value[q.putPos]
		q.value[q.putPos] = i
		q.putPos++
	} else {
		q.value[pos] = i
		q.putPos = 0
	}
	q.pos = pos
	//println("put", q.pos)
}
