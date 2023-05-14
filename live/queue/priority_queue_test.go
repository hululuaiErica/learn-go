package queue

import (
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	q1 := NewPriorityQueue[int](10, func(src int, dst int) int {
		if src > dst {
			return 1
		} else if src < dst {
			return -1
		} else {
			return 0
		}
	})

	q1.Enqueue(1)
	q1.Enqueue(2)
	q1.Enqueue(3)
	q1.Enqueue(4)

	q2 := NewPriorityQueue[int](10, func(src int, dst int) int {
		if src > dst {
			return -1
		} else if src < dst {
			return 1
		} else {
			return 0
		}
	})

	q2.Enqueue(1)
	q2.Enqueue(2)
	q2.Enqueue(3)
	q2.Enqueue(4)

	val1, _ := q1.Dequeue()
	val2, _ := q2.Dequeue()
	t.Log(val1, val2)
}
