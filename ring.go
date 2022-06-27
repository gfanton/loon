package main

import (
	"container/ring"
	"fmt"
	"sync"
)

type LoopFilter func(v interface{}) bool

type Loop struct {
	muRing *sync.RWMutex
	size   int
	root   *ring.Ring
}

type ForkedLoop Loop

func NewForkedLoop(l *Loop, size int) *ForkedLoop {
	fork := NewLoop(size)
	l.Do(func(n *ring.Ring) {
		fork.PushFrontValue(n)
	}, size)
	return (*ForkedLoop)(fork)

}

func NewLoop(size int) *Loop {
	ring := ring.New(size)
	return &Loop{
		muRing: &sync.RWMutex{},
		root:   ring,
		size:   size,
	}
}

func (l *Loop) Root() (r *ring.Ring) {
	l.muRing.Lock()
	r = l.root
	l.muRing.Unlock()
	return
}

func (r *Loop) fromRoot(root *ring.Ring) *Loop {
	return &Loop{
		muRing: r.muRing,
		size:   r.size,
		root:   root,
	}
}

func (r *Loop) PushFrontValue(v interface{}) {
	r.muRing.Lock()
	r.root.Value = v
	r.root = r.root.Next()
	r.muRing.Unlock()
}

func (r *Loop) PushBackValue(v interface{}) {
	r.muRing.Lock()
	r.root = r.root.Prev()
	r.root.Value = v
	r.muRing.Unlock()
}

func (r *Loop) DoPrevUntil(start *ring.Ring, fn func(n *ring.Ring) bool) (p *ring.Ring) {
	r.muRing.Lock()
	defer r.muRing.Unlock()

	if p = start; p != nil {
		fmt.Println("start:", p.Value)
		for p = start.Prev(); p != r.root; p = p.Prev() {
			if !fn(p) {
				return
			}
		}

		fn(p)
	}

	return
}

func (r *Loop) DoNextUntil(start *ring.Ring, fn func(n *ring.Ring) bool) (n *ring.Ring) {
	r.muRing.Lock()
	defer r.muRing.Unlock()

	if n = start; n != nil {
		if !fn(n) {
			return
		}
		for n = n.Next(); n != r.root; n = n.Next() {
			// fmt.Println("loop", n.Value)
			if !fn(n) {
				return
			}
		}
	}

	return
}

func (r *Loop) Do(fn func(n *ring.Ring), limit int) {
	r.muRing.Lock()

	if r.root != nil {
		p := r.root.Prev()
		for i := 0; p != r.root && i < limit; i++ {
			fn(p)
			p = p.Prev()
		}
		fn(p)
	}

	r.muRing.Unlock()
}

func (r *Loop) DoBackward(fn func(n *ring.Ring), limit int) {
	r.muRing.Lock()
	p := r.root.Next()
	for i := 0; i < limit && p != r.root; i++ {
		fn(p)
		p = r.root.Next()
	}
	r.muRing.Unlock()
}

func (r *Loop) Resize(n int) {
	r.muRing.Lock()
	switch {
	case n < r.size:
		skip := r.root
		for r.size = r.size + 1; r.size > n; r.size-- {
			skip = skip.Next()
		}

		r.root.Link(skip)

	case n > r.size:
		for ; r.size < n; r.size++ {
			r.root.Link(&ring.Ring{Value: nil})
		}
		r.size = n
	}
	r.muRing.Unlock()
}

type BufferWindow struct {
	muRing sync.RWMutex
	filter LoopFilter
	buffer *Buffer
	fork   *Loop
}

func NewBufferWindow(buffer *Buffer, filter LoopFilter, size int) *BufferWindow {
	fork := NewLoop(size)
	i := 0
	buffer.Loop().DoPrevUntil(buffer.Root(), func(n *ring.Ring) bool {
		fmt.Println("nvalue", n.Value)
		if i >= size || n.Value == nil {
			return false
		}

		i++
		fork.PushFrontValue(n)
		return true
	})

	return &BufferWindow{
		filter: filter,
		buffer: buffer,
		fork:   fork,
	}
}

func (w *BufferWindow) Size() (i int) {
	w.muRing.RLock()
	root := w.fork.Root()
	for c := root.Prev(); c.Value != nil && c != root; c = c.Prev() {
		i++
	}
	w.muRing.RUnlock()
	return
}

func (w *BufferWindow) DoFork(fn func(index int, v interface{})) {
	w.muRing.Lock()

	var i int
	w.fork.Do(func(r *ring.Ring) {
		var node interface{}
		if r.Value != nil {
			node = r.Value.(*ring.Ring).Value
		}
		fn(i, node)
		i++
	}, w.fork.size)

	w.muRing.Unlock()
}

func (w *BufferWindow) Update() {
	froot := w.fork.Root()
	broot := w.buffer.Root()
	if froot == nil {
		return
	}

	var cursor *ring.Ring
	if froot.Value != nil {
		cursor = froot.Value.(*ring.Ring)
	} else {
		cursor = broot
	}

	w.muRing.Lock()

	cp, cn := cursor, cursor.Next()

	var end *ring.Ring
	w.fork.DoNextUntil(froot, func(n *ring.Ring) bool {
		for ; cp != broot; cp = cp.Prev() {
			if w.filter(cp.Value) {
				fmt.Println("do prev nvalue", cp.Value, broot.Value)
				end, n.Value = n, cp
				cp = cp.Prev()
				return true
			}
		}

		if w.filter(cp.Value) {
			end, n.Value = n, cp
		}

		return false

	})

	w.fork.DoPrevUntil(froot, func(n *ring.Ring) bool {
		if n != end {
			for ; cn != broot; cn = cn.Next() {
				if w.filter(cn.Value) {
					fmt.Println("do next value", cn.Value, broot.Value)
					froot, n.Value = n, cn
					cn = cn.Next()
					return true
				}
			}

			froot, n.Value = n, nil
			return true
		}

		return false
	})

	w.fork.root = froot

	w.muRing.Unlock()
}

func (w *BufferWindow) Move(n int) {
	froot := w.fork.Root()
	broot := w.buffer.Root()
	loop := w.buffer.Loop()
	if froot == nil || broot == nil {
		fmt.Println(froot == nil)
		return
	}

	w.muRing.Lock()
	defer w.muRing.Unlock()

	if froot.Value == nil {
		prev := broot.Prev()
		if w.filter(prev.Value) {
			froot.Value = prev
		}
		return
	}

	switch {
	case n > 0:
		prev := froot.Prev()
		if prev.Value == nil {
			return
		}
		end := broot.Prev()
		start := prev.Value.(*ring.Ring)
		loop.DoPrevUntil(start, func(r *ring.Ring) bool {
			if r.Value == nil || r == end {
				return false
			}

			if w.filter(r.Value) {
				w.fork.PushFrontValue(r)
				n--
			}

			return n != 0
		})
	case n < 0:
		start := froot.Value.(*ring.Ring)
		loop.DoNextUntil(start.Next(), func(r *ring.Ring) bool {
			if r.Value == nil || r == broot {
				return false
			}

			if w.filter(r.Value) {
				w.fork.PushBackValue(r)
				n++
			}

			return n != 0
		})
	}
}
