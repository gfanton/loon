package main

import (
	"container/ring"
	"sync"
)

type Node struct {
	Line  Line
	Index uint
}

type Buffer struct {
	muRing sync.RWMutex
	loop   *Loop
	// size   int

	offset uint

	parser Parser
	reader Reader
}

func NewBuffer(size int, parser Parser, reader Reader) *Buffer {
	loop := NewLoop(size)
	return NewBufferFromRing(loop, parser, reader)
}

func NewBufferFromRing(ring *Loop, parser Parser, reader Reader) *Buffer {
	return &Buffer{
		loop: ring,
		// size:   size,
		parser: parser,
		reader: reader,
	}
}

func (r *Buffer) Infos() (lines, size int64, offset uint) {
	r.muRing.RLock()
	lines, size, offset = 0, int64(0), r.offset
	r.muRing.RUnlock()
	return
}

func (b *Buffer) Readline() (string, error) {
	line, err := b.reader.Readline()
	if err != nil {
		return "", err
	}

	b.addline(line)
	return line, nil
}

func (b *Buffer) Loop() (l *Loop) {
	b.muRing.RLock()
	l = b.loop
	b.muRing.RUnlock()
	return
}

func (b *Buffer) Lines() (l uint) {
	b.muRing.RLock()
	l = b.offset
	b.muRing.RUnlock()
	return
}

func (b *Buffer) Root() (r *ring.Ring) {
	b.muRing.RLock()
	r = b.loop.Root()
	b.muRing.RUnlock()
	return
}

func (b *Buffer) addline(line string) {
	b.muRing.Lock()

	b.offset++
	b.loop.PushFrontValue(&Node{
		Line:  b.parser.Parse(line),
		Index: b.offset,
	})

	b.muRing.Unlock()
}

// type Filter func(input, line string) bool

// var defaultFilter = func(input, line string) bool {
// 	inputs := strings.Split(input, " ")
// 	for _, in := range inputs {
// 		if strings.Contains(line, in) {
// 			return true
// 		}
// 	}

// 	return false
// }

// type Node struct {
// 	Content string
// 	Index   uint
// }

// type Buffer struct {
// 	muRing sync.RWMutex
// 	loop *Loop

// 	root   *ring.Ring
// 	cursor *ring.Ring
// 	offset uint

// 	// offset int64
// 	size int

// 	// subring *ring.Ring
// 	filter Filter
// 	input  string

// 	muRing sync.RWMutex

// 	parser Parser
// 	reader Reader
// }

// func NewBuffer(size int, parser Parser, reader Reader) *Buffer {
// 	ring := ring.New(size)
// 	return NewBufferFromRing(ring, parser, reader)
// }

// func NewBufferFromRing(ring *ring.Ring, parser Parser, reader Reader) *Buffer {
// 	size := ring.Len()
// 	return &Buffer{
// 		root:   ring,
// 		size:   size,
// 		parser: parser,
// 		reader: reader,
// 		filter: defaultFilter,
// 	}
// }

// func (r *Buffer) Infos() (lines, size int64, offset uint) {
// 	r.muRing.RLock()
// 	lines, size, offset = r.Lines(), int64(r.size), r.offset
// 	r.muRing.RUnlock()
// 	return
// }

// func (b *Buffer) ResetCursor() {
// 	b.muRing.Lock()
// 	b.cursor = nil
// 	b.offset = 0
// 	b.muRing.Unlock()
// }

// func (b *Buffer) MoveCursor(add, maxTopOffset int64) {
// 	var move int64

// 	b.muRing.Lock()
// 	defer b.muRing.Unlock()

// 	cursor := b.getCursor()
// 	rnode, ok := b.root.Prev().Value.(*Node)
// 	if !ok {
// 		return
// 	}

// 	maxoffset := rnode.Index - uint(maxTopOffset)
// 	// offset := b.offset
// 	for i := 0; add != 0; i++ {
// 		if add > 0 {
// 			cursor = cursor.Prev()
// 			move = -1
// 		} else {
// 			cursor = cursor.Next()
// 			move = 1
// 		}

// 		if cursor == b.root || cursor.Value == nil {
// 			log.Printf("root: %t, isNil: %t", cursor == b.root, cursor.Value == nil)

// 			// cursor = nil
// 			// b.offset = 0
// 			return
// 		}

// 		node := cursor.Value.(*Node)

// 		log.Printf("found: %d, cursor: %s, offset: %d, max: %d", add, node.Content, rnode.Index-node.Index, maxoffset)

// 		if b.filter(b.input, node.Content) {
// 			add = add + move
// 		} else {
// 			maxoffset--
// 		}

// 		if rnode.Index-node.Index >= maxoffset {
// 			return
// 		}

// 		b.cursor = cursor
// 		// if line := cursor.Value.(string); !b.filter(b.input, line) {
// 		// 	continue
// 		// }

// 		// offset = offset - move
// 		// if offset < 0 {
// 		// 	b.cursor = nil
// 		// 	b.offset = 0
// 		// 	return
// 		// }

// 		// if offset > int64(maxoffset) {
// 		// 	return
// 		// }
// 	}

// 	// b.offset = offset
// }

// func (b *Buffer) Clear() {
// 	b.muRing.Lock()
// 	b.reader.ResetLines()
// 	b.root.Value = nil
// 	b.root = b.root.Next()
// 	b.cursor = b.root
// 	b.muRing.Unlock()
// }

// func (r *Buffer) SetFilter(filter Filter) {
// 	r.muRing.Lock()
// 	r.filter = filter
// 	r.cursor = nil
// 	r.muRing.Unlock()
// }

// func (b *Buffer) GetCursor() (r *ring.Ring) {
// 	b.muRing.Lock()
// 	r = b.getCursor()
// 	b.muRing.Unlock()
// 	return
// }

// func (b *Buffer) getCursor() *ring.Ring {
// 	if b.cursor != nil {
// 		return b.cursor
// 	}

// 	return b.root.Prev()
// }

// func (b *Buffer) GetRoot() (root *ring.Ring) {
// 	b.muRing.Lock()
// 	root = b.root.Prev()
// 	b.muRing.Unlock()
// 	return
// }

// type Matchs struct {
// 	lines           []Line
// 	size, maxoffset int
// 	index           int
// }

// func (b *Buffer) FilterLines(input string, limit int) *Matchs {
// 	lines := make([]Line, limit)

// 	var maxoffset, bn, fn int

// 	b.muRing.Lock()

// 	b.input = input
// 	cursor := b.getCursor()
// 	for p := cursor; p != b.root && bn < limit; p = p.Prev() {
// 		if p.Value == nil {
// 			break
// 		}

// 		line := p.Value.(*Node).Content
// 		if b.filter(b.input, line) {
// 			pline := b.parser.Parse(line)
// 			lines[bn] = pline
// 			bn++
// 			if pline.Len() > maxoffset {
// 				maxoffset = pline.Len()
// 			}
// 		}
// 	}

// 	if bn < limit {
// 		forwardLimit := limit - bn
// 		forwardline := make([]Line, forwardLimit)
// 		for p := cursor.Next(); p != b.root && fn < forwardLimit; p = p.Next() {
// 			if p.Value == nil {
// 				break
// 			}

// 			line := p.Value.(*Node).Content
// 			if b.filter(b.input, line) {
// 				pline := b.parser.Parse(line)
// 				fn++
// 				forwardline[forwardLimit-fn] = pline
// 				if pline.Len() > maxoffset {
// 					maxoffset = pline.Len()
// 				}
// 			}
// 		}

// 		lines = append(forwardline[forwardLimit-fn:], lines[:limit-fn]...)
// 	}

// 	b.muRing.Unlock()

// 	log.Printf("lines: %v", lines)
// 	return &Matchs{
// 		lines:     lines,
// 		size:      bn + fn,
// 		maxoffset: maxoffset,
// 	}

// }

// func (b *Buffer) Readline() (string, error) {
// 	line, err := b.reader.Readline()
// 	if err != nil {
// 		return "", err
// 	}
// 	b.addline(line)
// 	return line, nil
// }

// func (b *Buffer) addline(line string) {
// 	b.muRing.Lock()
// 	if b.root.Value == nil {
// 		b.root.Value = &Node{}
// 	}

// 	b.offset++

// 	node := b.root.Value.(*Node)
// 	node.Content = line
// 	node.Index = b.offset

// 	b.root = b.root.Next()

// 	b.muRing.Unlock()
// }

// func (b *Buffer) CursorOffset() (o uint) {
// 	b.muRing.RLock()
// 	o = b.offset
// 	b.muRing.RUnlock()
// 	return
// }

// func (b *Buffer) Lock() {
// 	b.muRing.Lock()
// 	if b.cursor == nil {
// 		b.cursor = b.root.Prev()
// 	}
// 	b.muRing.Unlock()
// }

// func (b *Buffer) Lines() int64 {
// 	return b.reader.Lines()
// }
