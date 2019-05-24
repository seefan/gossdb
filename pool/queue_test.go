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
		{"1", fields{3, 0, []int{1, 3, 5, 6, 8}, 5}, 6},
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
		{"1", fields{3, 1, []int{1, 3, 5, 6, 8}, 5}, args{7}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				pos:    tt.fields.pos,
				putPos: tt.fields.putPos,
				value:  tt.fields.value,
				size:   tt.fields.size,
			}
			q.Put(tt.args.i)
		})
	}
}
func TestQueue_All(t *testing.T) {
	q := newQueue(10)

	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)
	t.Log(q.Pop())
	t.Log(q.value)

}

func BenchmarkQueue10(b *testing.B) {
	b.SetParallelism(10)

	q := newQueue(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			re := q.Pop()
			if re != -1 {
				q.Put(re)
			}

		}
	})
}
func BenchmarkQueue100(b *testing.B) {
	b.SetParallelism(100)

	q := newQueue(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			re := q.Pop()
			if re != -1 {
				q.Put(re)
			}

		}
	})
}
func BenchmarkQueue1000(b *testing.B) {
	b.SetParallelism(1000)

	q := newQueue(100)

	failed := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			re := q.Pop()
			if re != -1 {
				q.Put(re)
			} else {
				failed++
			}

		}
	})
	b.Log("fail", failed)
}

func BenchmarkMap(b *testing.B) {
	b.SetParallelism(100)

	m := &sync.Map{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store(1, 1)
			_, ok := m.Load(1)
			if !ok {
				b.Error(ok)
			}
		}
	})
}
