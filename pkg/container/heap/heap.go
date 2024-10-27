package heap

import (
	"container/heap"
	"errors"
)

type heapEle[T any] struct {
	Element T
	Score   float64
}

// Heap is a max-heap
type Heap[T any] []heapEle[T]

// Len is the number of elements in the collection.
func (h Heap[T]) Len() int { return len(h) }

// Less reports whether the element with index i should sort before the element with index j.
func (h Heap[T]) Less(i, j int) bool { return h[i].Score < h[j].Score }

// Swap swaps the elements with indexes i and j.
func (h Heap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push adds x as element Len() and returns the new length of the heap.
func (h *Heap[T]) Push(x interface{}) {
	*h = append(*h, x.(heapEle[T]))
}

// Pop removes and returns the maximum element from the heap.
func (h *Heap[T]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

var ErrHeapEmpty = errors.New("heap is empty")
var ErrHeapFull = errors.New("heap is full")

type MinHeap[T any] interface {
	Push(t T, score float64) error
	Pop() (T, float64, error)
	Len() int
}

type localMinHeap[T any] struct {
	h    *Heap[T]
	len  int
	size int
}

func (l *localMinHeap[T]) Push(t T, score float64) error {
	if l.len == l.size {
		return ErrHeapFull
	}
	heap.Push(l.h, heapEle[T]{
		Element: t,
		Score:   score,
	})
	l.len++
	return nil
}

func (l *localMinHeap[T]) Pop() (T, float64, error) {
	if l.len == 0 {
		var t T
		return t, 0.0, ErrHeapEmpty
	}
	ele := heap.Pop(l.h)
	l.len--
	return ele.(heapEle[T]).Element, ele.(heapEle[T]).Score, nil
}

func (l *localMinHeap[T]) Len() int {
	return l.len
}

func NewLocalMinHeap[T any](size int) MinHeap[T] {
	l := &localMinHeap[T]{len: 0, size: size}
	h := make(Heap[T], 0, size)
	heap.Init(&h)
	l.h = &h
	return l
}
