package main

import (
	"container/ring"
	"strings"
	"sync"
)

type Filter func(input, line string) bool

var defaultFilter = func(input, line string) bool {
	inputs := strings.Split(input, " ")
	for _, in := range inputs {
		if strings.Contains(line, in) {
			return true
		}
	}

	return false
}

type Buffer struct {
	root   *ring.Ring
	cursor *ring.Ring

	offset int64
	size   int

	// subring *ring.Ring
	filter Filter
	input  string

	muRing sync.RWMutex

	parser Parser
	reader Reader
}

func NewBuffer(size int, parser Parser, reader Reader) *Buffer {
	ring := ring.New(size + 1)
	return NewBufferFromRing(ring, parser, reader)
}

func NewBufferFromRing(ring *ring.Ring, parser Parser, reader Reader) *Buffer {
	size := ring.Len()
	return &Buffer{
		root:   ring,
		size:   size,
		parser: parser,
		reader: reader,
		filter: defaultFilter,
	}
}

func (r *Buffer) Infos() (lines, size, offset int64) {
	r.muRing.RLock()
	lines, size, offset = r.Lines(), int64(r.size), r.offset
	r.muRing.RUnlock()
	return
}

func (b *Buffer) ResetCursor() {
	b.muRing.Lock()
	b.cursor = nil
	b.offset = 0
	b.muRing.Unlock()
}

func (b *Buffer) MoveCursor(maxoffset, add int64) {
	var move int64

	b.muRing.Lock()
	defer b.muRing.Unlock()

	cursor := b.getCursor()
	offset := b.offset
	for add != 0 {
		if add > 0 {
			cursor = cursor.Prev()
			move = -1
		} else {
			cursor = cursor.Next()
			move = 1
		}

		if cursor == b.root || cursor.Value == nil {
			cursor = nil
			b.offset = 0
			return
		}

		if line := cursor.Value.(string); !b.filter(b.input, line) {
			continue
		}

		add = add + move

		offset = offset - move
		if offset < 0 {
			b.cursor = nil
			b.offset = 0
			return
		}

		if offset > int64(maxoffset) {
			return
		}

		b.offset = offset
		b.cursor = cursor
	}
}

func (b *Buffer) Clear() {
	b.muRing.Lock()
	b.reader.ResetLines()
	b.root.Value = nil
	b.root = b.root.Next()
	b.cursor = b.root
	b.muRing.Unlock()
}

func (r *Buffer) SetFilter(filter Filter) {
	r.muRing.Lock()
	r.filter = filter
	r.cursor = nil
	r.muRing.Unlock()
}

func (b *Buffer) GetCursor() (r *ring.Ring) {
	b.muRing.Lock()
	r = b.getCursor()
	b.muRing.Unlock()
	return
}

func (b *Buffer) getCursor() *ring.Ring {
	if b.cursor != nil {
		return b.cursor
	}

	return b.root.Prev()
}

func (b *Buffer) GetRoot() *ring.Ring {
	return b.root.Prev()
}

type Matchs struct {
	lines           []Line
	size, maxoffset int
	index           int
}

func (b *Buffer) FilterLines(input string, limit int) *Matchs {
	lines := make([]Line, limit)

	var maxoffset, bn, fn int

	b.muRing.Lock()

	b.input = input
	cursor := b.getCursor()
	for p := cursor; p != b.root && bn < limit; p = p.Prev() {
		if p.Value == nil {
			break
		}

		line := p.Value.(string)
		if b.filter(b.input, line) {
			pline := b.parser.Parse(line)
			lines[bn] = pline
			bn++
			if pline.Len() > maxoffset {
				maxoffset = pline.Len()
			}
		}
	}

	if bn < limit {
		forwardLimit := limit - bn
		forwardline := make([]Line, forwardLimit)
		for p := cursor.Next(); p != b.root && fn < forwardLimit; p = p.Next() {
			if p.Value == nil {
				break
			}

			line := p.Value.(string)
			if b.filter(b.input, line) {
				pline := b.parser.Parse(line)
				fn++
				forwardline[forwardLimit-fn] = pline
				if pline.Len() > maxoffset {
					maxoffset = pline.Len()
				}
			}
		}

		lines = append(forwardline[forwardLimit-fn:], lines[:limit-fn]...)
	}

	b.muRing.Unlock()

	// log.Printf("lines: %v", lines)
	return &Matchs{
		lines:     lines,
		size:      bn + fn,
		maxoffset: maxoffset,
	}

}

func (b *Buffer) Readline() (string, error) {
	line, err := b.reader.Readline()
	if err != nil {
		return "", err
	}

	b.muRing.Lock()
	b.root.Value = line
	b.root = b.root.Next()
	if b.cursor != nil && b.offset < int64(b.size) && b.filter(b.input, line) {
		b.offset++
	}

	b.muRing.Unlock()

	return line, nil
}

func (b *Buffer) CursorOffset() (o int64) {
	b.muRing.RLock()
	o = b.offset
	b.muRing.RUnlock()
	return
}

func (b *Buffer) Lock() {
	b.muRing.Lock()
	if b.cursor == nil {
		b.cursor = b.root.Prev()
	}
	b.muRing.Unlock()
}

func (b *Buffer) Lines() int64 {
	return b.reader.Lines()
}
