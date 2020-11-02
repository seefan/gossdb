package queue

import (
	"testing"
)

func Test_queue(t *testing.T) {

}

func BenchmarkQueues(b *testing.B) {
	b.ResetTimer()

	re := []int{1, 2, 3, 4, 5, 6}
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			re = re[1:]
		} else {
			re = append(re, 1)
		}
	}
}
func BenchmarkQueueq(b *testing.B) {
	b.ResetTimer()

	q := NewQueue(100)
	var re int
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			re = q.Pop()
		} else {
			q.Put(re)
		}
	}
}

func TestQueue_Pop(t *testing.T) {
	q := NewQueue(5)
	vs := []int{4, 3, 2, 1, 0, -1}
	for i := 0; i < 6; i++ {
		pos := q.Pop()
		t.Log(pos, q.value)
		if pos != vs[i] {
			t.Error("pop err")
		}
	}
	for i := 0; i < 6; i++ {
		pos := q.Put(i)
		t.Log(pos, q.value)
	}
}
