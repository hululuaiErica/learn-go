package queue

import "testing"

func TestChangePriority(t *testing.T) {
	q := NewPriorityQueue[*Element](100,
		func(src *Element, dst *Element) int {
			if src.Priority > dst.Priority {
				return 1
			} else if src.Priority < dst.Priority {
				return -1
			}
			return 0
		})
	e1 := &Element{
		Data:     10,
		Priority: 200,
	}
	_ = q.Enqueue(e1)
	e2 := &Element{
		Data:     10,
		Priority: 100,
	}
	_ = q.Enqueue(e2)
	//e1.Priority = 10
	val, _ := q.Dequeue()
	println(val)
}

type Element struct {
	Data     any
	Priority int
}
