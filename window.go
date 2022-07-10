package main

import (
	"container/ring"
	"sync"
)

type Window[T any] struct {
	muRing sync.RWMutex

	head, tail   *ring.Ring
	size, length int
	emptyT       T
}

func NewWindow[T any](size int) *Window[T] {
	ring := ring.New(size)
	return &Window[T]{
		head: ring,
		tail: ring,
		size: size,
	}
}

func (w *Window[T]) IsFull() (yes bool) {
	w.muRing.RLock()
	if w.size > 0 {
		yes = w.head.Value != nil && w.tail.Value != nil
	}
	w.muRing.RUnlock()
	return
}

func (w *Window[T]) IsEmpty() (yes bool) {
	w.muRing.RLock()

	if w.size > 0 {
		yes = w.head.Value == nil && w.tail.Value == nil
	}
	w.muRing.RUnlock()
	return
}

func (w *Window[T]) Size() (size, length int) {
	w.muRing.RLock()
	size, length = w.size, w.length
	w.muRing.RUnlock()
	return
}

func (w *Window[T]) Reset() (v T) {
	w.muRing.RLock()

	if w.size > 0 {
		DoRingNext(w.tail, func(r *ring.Ring) bool {
			r.Value = nil
			return true
		})
		w.head = w.tail
		w.length = 0
	}

	w.muRing.RUnlock()
	return
}

func (w *Window[T]) HeadValue() (value T) {
	w.muRing.RLock()

	if w.size > 0 {
		if prev := w.head.Prev(); prev.Value != nil {
			value = prev.Value.(T)
		}
	}

	w.muRing.RUnlock()
	return
}

func (w *Window[T]) TailValue() (value T) {
	w.muRing.RLock()

	if w.size > 0 {
		if w.tail.Value != nil {
			value = w.tail.Value.(T)
		}
	}

	w.muRing.RUnlock()
	return
}

func (w *Window[T]) Do(f func(v T)) {
	w.muRing.RLock()

	DoRingNext(w.tail, func(r *ring.Ring) bool {
		if r.Value == nil {
			return false
		}

		f(r.Value.(T))
		return true
	})

	w.muRing.RUnlock()
}

func (w *Window[T]) PushFront(v T) {
	w.muRing.Lock()

	if w.size > 0 {
		if w.head.Value != nil && w.tail.Value != nil {
			w.tail = w.tail.Next()
		} else {
			w.length++
		}

		w.head.Value = v
		w.head = w.head.Next()
	}

	w.muRing.Unlock()
}

func (w *Window[T]) PushBack(v T) {
	w.muRing.Lock()

	if w.size > 0 {
		if w.head.Value != nil && w.tail.Value != nil {
			w.head = w.head.Prev()
		} else {
			w.length++
		}

		w.tail = w.tail.Prev()
		w.tail.Value = v
	}

	w.muRing.Unlock()
}

func (w *Window[T]) SlideFront() {
	w.muRing.Lock()

	if w.length > 0 {
		w.tail.Value = nil
		w.tail = w.tail.Next()
		w.length--
	}

	w.muRing.Unlock()
}

func (w *Window[T]) SlideBack() {
	w.muRing.Lock()

	if w.length > 0 {
		w.head = w.head.Prev()
		w.head.Value = nil
		w.length--
	}

	w.muRing.Unlock()
}

func (w *Window[T]) Slice() []T {
	w.muRing.RLock()

	lines := make([]T, w.length)
	i := 0
	DoRingNext(w.tail, func(r *ring.Ring) bool {
		if i >= w.length {
			return false
		}

		lines[i] = r.Value.(T)
		i++
		return true
	})

	w.muRing.RUnlock()
	return lines
}

func (w *Window[T]) Resize(n int) bool {
	w.muRing.Lock()
	defer w.muRing.Unlock()

	switch {
	case n < w.size:
		skip := w.head
		for w.size = w.size + 1; w.size > n; w.size-- {
			skip = skip.Next()
		}

		w.head.Link(skip)

	case n > w.size:
		if w.head == nil {
			w.head = &ring.Ring{Value: nil}
			w.size++
		}

		if w.tail == nil {
			w.tail = w.head
		}

		for ; w.size < n; w.size++ {
			w.head.Link(&ring.Ring{Value: nil})
		}
		w.size = n
	default:
		return false
	}

	if w.length > w.size {
		w.length = w.size
	}

	return true
}
