/*
@Time : 2019-04-30 20:26
@Author : seefan
@File : block
@Software: gossdb
*/
package gossdb

// 使用slice实现的池
type Block struct {
	//id
	index int
	//长度
	length int
	//available
	available bool
	// last error
	lastError error
	//ava id
	queue *Queue
}

//附加连接
//
// cs [] 附加的连接
// 返回 error 可能的错误
func newBlock(index, size int) *Block {
	s := &Block{
		index:  index,
		length: size,
		queue:  newQueue(size),
	}

	for i := 0; i < s.length; i++ {
		s.queue.Put(s.index + i)
	}
	return s
}

//获取一个连接
//
// 返回 *PooledClient 缓存的连接
// 返回 error 可能的错误
func (s *Block) Get() int {
	return s.queue.Pop()
}

//回收连接
//
// element *PooledClient 要回收的连接
func (s *Block) Set(i int) {
	s.queue.Put(i)
}
