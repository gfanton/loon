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

func (b *BufferWindow[T]) Readline() (line string, err error) {
	if line, err = b.reader.Readline(); err == nil {
		b.mu.Lock()

		// top := b.buffer.Head() == b.window.HeadValue()

		// fmt.Printf("%+v\n", line)

		pline := b.parser.Parse(line)
		n := b.buffer.AddValue(pline)

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

func (b *BufferWindow[T]) Size() (size, length int) {
	b.mu.Lock()
	size, length = b.window.Size()
	b.mu.Unlock()
	return
}

func (b *BufferWindow[T]) Resize(n int) {
	b.mu.Lock()
	b.window.Resize(n)
	b.mu.Unlock()
}

func (b *BufferWindow[T]) Lock(yes bool) {
	b.mu.Lock()
	b.lock = yes
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
			// fmt.Printf("from: %v -> down: %v\n", root.Value, r.Value)
			switch {
			case r == bufferHead:
				return false
			case r.Value == nil:
				return false
			case b.filterRing(r):
				b.follow = false
				b.window.PushBack(r)
				n++
			}

			return n != 0
		})

	case n > 0:
		// head := b.window.HeadValue()
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
				// fmt.Printf("push %+v\n", r.Value)
				b.window.PushFront(r)
				n--
			}

			return n != 0 && r != bufferHead
		})
	}
}

func (b *BufferWindow[T]) Sync() {
	b.mu.Lock()

	bufferHead := b.buffer.Head()
	windowHead := b.window.HeadValue()

	if windowHead == nil {
		windowHead = bufferHead
	}

	b.window.Reset()

	var wg sync.WaitGroup
	// var mu sync.Mutex

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

		// var full bool
		// DoRingNext(windowHead, func(r *ring.Ring) bool {
		// 	if r.Value == nil {
		// 		return false
		// 	}

		// 	if ok := b.filterRing(r); ok {
		// 		if !full {
		// 			b.window.PushFront(r)
		// 		}
		// 		if r == bufferHead {
		// 			b.follow = true
		// 		}
		// 	}

		// 	full = b.window.IsFull()
		// 	return (!full || !b.follow) && r != bufferHead

		// })
		wg.Done()
	}()

	wg.Wait()

	b.mu.Unlock()
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
	b.window.Do(func(r *ring.Ring) {
		if i >= l {
			return
		}

		slice[i] = r.Value.(T)
		i++
	})

	return slice
}

func (b *BufferWindow[T]) Lines() (l uint) {
	b.mu.Lock()
	l = b.buffer.Lines()
	b.mu.Unlock()
	return
}

func (b *BufferWindow[T]) filterRing(r *ring.Ring) (ok bool) {
	return b.filter(r.Value.(T))
}
