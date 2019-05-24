//  @time : 2019-05-24 15:06
//  @author : seefan
//  @file : ring
//  @software: gossdb

package pool

import (
	"runtime"
	"sync/atomic"
)

//Ring ring slice
type Ring struct {
	value       []int32
	read        uint32
	write       uint32
	maxRead     uint32
	available   int32
	capacity    uint32
	capacityMod uint32
}

func newRing(vss int) *Ring {
	vs := make([]int, vss)
	for i := 0; i < vss; i++ {
		vs[i] = i
	}
	l := uint32(vss)
	size := minQuantity(l)
	if l == size {
		size = l * 2
	}
	r := &Ring{
		capacity:    size,
		capacityMod: size - 1,
		value:       make([]int32, size),
		maxRead:     l,
		write:       l,
	}
	for i, v := range vs {
		r.value[i] = int32(v)
	}
	return r
}

//Pop get a index
//
//  @return int index
//
//获取一个可以连接的位置
func (r *Ring) Pop() int {
	read := atomic.LoadUint32(&r.read)
	max := atomic.LoadUint32(&r.maxRead)
	if read&r.capacityMod == max&r.capacityMod {
		return -1
	}
	next := read + 1
	v := atomic.LoadInt32(&r.value[read&r.capacityMod])
	if atomic.CompareAndSwapUint32(&r.read, read, next) {
		atomic.AddInt32(&r.available, -1)
		return int(v)
	}
	return -1
}

//Put return a index
//
//  @param i index
//  @return int pos
//
//归还索引值
func (r *Ring) Put(v int) {
	var read, write, next, max uint32
	for {
		max = atomic.LoadUint32(&r.maxRead)
		write = atomic.LoadUint32(&r.write)
		if max != write { //不一致,表明有写请求尚未完成.这意味着,有写请求成功申请了空间但数据还没完全写进队列.所以如果有线程要读取,必须要等到写线程将数完全据写入到队列之后.
			runtime.Gosched()
			continue
		}
		read = atomic.LoadUint32(&r.read)
		next = write + 1
		if read&r.capacityMod == next&r.capacityMod { //空间满了
			runtime.Gosched()
			continue
		}
		if atomic.CompareAndSwapUint32(&r.write, write, next) {
			atomic.StoreInt32(&r.value[write&r.capacityMod], int32(v))
			atomic.StoreUint32(&r.maxRead, next)
			atomic.AddInt32(&r.available, 1)
			break
		}
		runtime.Gosched()
	}
}
func minQuantity(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

//Available queue available size
func (r *Ring) Available() int {
	c := atomic.LoadInt32(&r.available)
	return int(c)
}
