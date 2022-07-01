package main

import (
	"container/ring"
	"sync"
)

type Buffer[T any] struct {
	size int
	ring *ring.Ring

	offset uint
	muRing sync.RWMutex
}

func NewBuffer[T any](size int) *Buffer[T] {
	if size == 0 {
		panic("cannot use 0 sized buffer")
	}

	ring := ring.New(size)
	return &Buffer[T]{ring: ring, size: size}
}

func (b *Buffer[T]) AddValue(value T) (node *ring.Ring) {
	b.muRing.Lock()
	b.offset++
	b.ring.Value = value

	node = b.ring

	b.ring = b.ring.Next()
	b.muRing.Unlock()
	return
}

func (b *Buffer[T]) DoPrev(f func(n *ring.Ring, v T) bool) {
	b.muRing.RLock()

	DoRingPrev(b.ring, func(r *ring.Ring) bool {
		if r.Value == nil {
			return false
		}

		return f(r, r.Value.(T))
	})

	b.muRing.RUnlock()
}

func (b *Buffer[T]) Lines() (i uint) {
	b.muRing.RLock()
	i = b.offset
	b.muRing.RUnlock()
	return
}

func (b *Buffer[T]) Full() (yes bool) {
	b.muRing.RLock()
	yes = b.ring.Value != nil
	b.muRing.RUnlock()
	return
}

func (b *Buffer[T]) Head() (head *ring.Ring) {
	b.muRing.RLock()
	if prev := b.ring.Prev(); prev != nil {
		head = prev
	}
	b.muRing.RUnlock()
	return
}
