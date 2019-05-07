/*
@Time : 2019-04-30 20:42
@Author : seefan
@File : queue
@Software: gossdb
*/
package pool

import (
	"reflect"
	"sync"
	"testing"
)

func Test_newQueue(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want *Queue
	}{
		{"1", args{10}, newQueue(10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newQueue(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_Pop(t *testing.T) {
	type fields struct {
		pos    int
		putPos int
		value  []int
		size   int
	}
	tests := []struct {
		name   string
		fields fields
		wantRe int
	}{
		{"1", fields{2, 0, []int{1, 3, 5, 6, 8}, 5}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				pos:    tt.fields.pos,
				putPos: tt.fields.putPos,
				value:  tt.fields.value,
				size:   tt.fields.size,
			}
			if gotRe := q.Pop(); gotRe != tt.wantRe {
				t.Errorf("Queue.Pop() = %v, want %v", gotRe, tt.wantRe)
			}
		})
	}
}

func TestQueue_Put(t *testing.T) {
	type fields struct {
		pos    int
		putPos int
		value  []int
		size   int
	}
	type args struct {
		i int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{"1", fields{2, 0, []int{1, 3, 5, 6, 8}, 5}, args{1}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				pos:    tt.fields.pos,
				putPos: tt.fields.putPos,
				value:  tt.fields.value,
				size:   tt.fields.size,
			}
			if got := q.Put(tt.args.i); got != tt.want {
				t.Errorf("Queue.Put() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestQueue_All(t *testing.T) {
	q := newQueue(10)
	for i := 0; i < 10; i++ {
		q.Add(i + 10)
	}
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Put(20))
	t.Log(q.value)
	t.Log(q.Put(21))
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Put(23))
	t.Log(q.value)
	t.Log(q.Put(24))
	t.Log(q.value)
	t.Log(q.Put(25))
	t.Log(q.value)
	t.Log(q.Put(26))
	t.Log(q.value)
	t.Log(q.Put(27))
	t.Log(q.value)
	t.Log(q.Put(28))
}

func BenchmarkQueue10(b *testing.B) {
	b.SetParallelism(10)
	var lock sync.Mutex
	q := newQueue(100)
	for i := 0; i < 10; i++ {
		q.Add(i + 10)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock.Lock()
			re := q.Pop()
			if re != -1 {
				q.Put(re)
			}
			lock.Unlock()
		}
	})
}
func BenchmarkQueue100(b *testing.B) {
	b.SetParallelism(100)
	var lock sync.Mutex
	q := newQueue(100)
	for i := 0; i < 10; i++ {
		q.Add(i + 10)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock.Lock()
			re := q.Pop()
			if re != -1 {
				q.Put(re)
			}
			lock.Unlock()
		}
	})
}
func BenchmarkQueue1000(b *testing.B) {
	b.SetParallelism(50000)
	var lock sync.Mutex
	q := newQueue(100)
	for i := 0; i < 100; i++ {
		q.Add(i + 10)
	}
	failed := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock.Lock()
			re := q.Pop()
			if re != -1 {
				q.Put(re)
			} else {
				failed++
			}
			lock.Unlock()
		}
	})
	b.Log("fail", failed)
}
