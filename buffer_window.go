package main

import (
	"container/ring"
	"sync"
)

type WindowRing = Window[*ring.Ring]

type Filter[T any] func(value T) bool

var NoFilter = func(v any) bool {
	return v != nil
}

type BufferWindow[T any] struct {
	reader Reader
	filter Filter[T]
	parser Parser[T]
	buffer *Buffer[T]

	window       *WindowRing
	mu           sync.Mutex
	lock, follow bool

	// for test purpose
	sync bool
}

type BufferWindowOptions[T any] struct {
	Reader Reader
	Filter Filter[T]
	Parser Parser[T]
	Buffer *Buffer[T]
}

func NewBufferWindow[T any](size int, opts *BufferWindowOptions[T]) *BufferWindow[T] {
	window := NewWindow[*ring.Ring](size)
	return &BufferWindow[T]{
		filter: opts.Filter,
		reader: opts.Reader,
		parser: opts.Parser,
		buffer: opts.Buffer,
		follow: true,
		lock:   false,
		window: window,
	}
}

func (b *BufferWindow[T]) Readline() (value T, err error) {
	var sid SourceID
	var line string
	if line, sid, err = b.reader.Readline(); err == nil {
		b.mu.Lock()

		value = b.parser.Parse(sid, line)
		n := b.buffer.AddValue(value)

		switch {
		case b.window.IsEmpty():
			if b.filterRing(n) {
				b.window.PushFront(n)
			}
		case n == b.window.TailValue(): // buffer has catch windows tail
			b.window.SlideFront()
			fallthrough
		case !b.lock && b.follow, !b.window.IsFull():
			b.moveFrom(b.window.HeadValue(), 1)
		}

		b.mu.Unlock()
	}

	return
}

func (b *BufferWindow[T]) WindowSize() (size, length int) {
	b.mu.Lock()
	size, length = b.window.Size()
	b.mu.Unlock()
	return
}

func (b *BufferWindow[T]) Resize(n int) {
	b.mu.Lock()
	if b.window.Resize(n) {
		b.refresh()
	}
	b.mu.Unlock()
}

func (b *BufferWindow[T]) Lock(yes bool) {
	b.mu.Lock()
	b.lock = yes
	b.mu.Unlock()
}

func (b *BufferWindow[T]) Refresh() {
	b.mu.Lock()
	b.refresh()
	b.mu.Unlock()
}

func (b *BufferWindow[T]) MoveFront() {
	b.mu.Lock()
	root := b.window.HeadValue()
	if root == nil {
		root = b.buffer.Head()
	}
	b.moveFrom(root, b.buffer.Size())
	b.mu.Unlock()
}

func (b *BufferWindow[T]) MoveBack() {
	b.mu.Lock()
	root := b.window.TailValue()
	if root == nil {
		root = b.buffer.Head()
	}
	b.moveFrom(root, -b.buffer.Size())
	b.mu.Unlock()
}

func (b *BufferWindow[T]) Clear() {
	b.mu.Lock()
	b.buffer.Reset()
	b.refresh()
	b.mu.Unlock()
}

func (b *BufferWindow[T]) Move(n int) {
	b.mu.Lock()

	var root *ring.Ring
	switch {
	case n < 0:
		if root = b.window.TailValue(); root == nil {
			root = b.buffer.Head()
		}
	case n > 0:
		if root = b.window.HeadValue(); root == nil {
			root = b.buffer.Head()
		}
	}

	b.moveFrom(root, n)

	b.mu.Unlock()
}

func (b *BufferWindow[T]) moveFrom(root *ring.Ring, n int) {
	bufferHead := b.buffer.Head()
	if bufferHead == nil {
		return
	}

	switch {
	case n < 0:
		DoRingPrev(root, func(r *ring.Ring) bool {
			switch {
			case r == bufferHead, r.Value == nil:
				return false
			case b.filterRing(r):
				b.follow = false
				b.window.PushBack(r)
				n++
			}

			return n != 0
		})

	case n > 0:
		DoRingNext(root, func(r *ring.Ring) bool {
			if r == bufferHead {
				b.follow = true
			}

			switch {
			case r == root:
				return r != bufferHead
			case r.Value == nil:
				return false
			case b.filterRing(r):
				b.window.PushFront(r)
				n--
			}

			return n != 0 && r != bufferHead
		})
	}
}

func (b *BufferWindow[T]) refresh() {

	bufferHead := b.buffer.Head()
	windowHead := b.window.HeadValue()

	if windowHead == nil {
		windowHead = bufferHead
	}

	b.window.Reset()

	var wg sync.WaitGroup

	b.follow = false

	wg.Add(1)
	go func() {
		DoRingPrev(windowHead, func(r *ring.Ring) bool {
			if r == bufferHead || r.Value == nil {
				return false
			}

			if ok := b.filterRing(r); ok {
				b.window.PushBack(r)
			}

			return !b.window.IsFull()
		})
		wg.Done()
	}()

	if b.sync { // for testing purpose
		wg.Wait()
	}

	wg.Add(1)
	go func() {
		DoRingNext(windowHead, func(r *ring.Ring) bool {
			if r.Value == nil {
				return false
			}

			if ok := b.filterRing(r); ok {
				b.window.PushFront(r)
				if r == bufferHead {
					b.follow = true
				}
			}

			return !b.window.IsFull() && r != bufferHead

		})
		wg.Done()
	}()

	wg.Wait()
}

func (b *BufferWindow[T]) Lines() (l uint) {
	b.mu.Lock()
	l = b.buffer.Lines()
	b.mu.Unlock()
	return
}

func (b *BufferWindow[T]) Do(f func(i int, v T) bool) (size int) {
	b.mu.Lock()

	_, l := b.window.Size()

	b.window.Do(func(r *ring.Ring) (ok bool) {
		if size >= l {
			return false
		}

		ok = f(size, r.Value.(T))
		size++
		return
	})

	b.mu.Unlock()

	return size

}

func (b *BufferWindow[T]) Slice() (slice []T) {
	b.mu.Lock()
	slice = b.slice()
	b.mu.Unlock()
	return
}

func (b *BufferWindow[T]) slice() []T {
	_, l := b.window.Size()
	slice := make([]T, l)

	i := 0
	b.window.Do(func(r *ring.Ring) bool {
		if i >= l {
			return false
		}

		slice[i] = r.Value.(T)
		i++
		return true
	})

	return slice
}

func (b *BufferWindow[T]) filterRing(r *ring.Ring) (ok bool) {
	return b.filter(r.Value.(T))
}
