package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type position struct {
	mu sync.RWMutex
	s  tcell.Screen

	// muCursor, muOffset   sync.RWMutex

	window               int64
	cursor, offset       int64
	maxOffset, maxCursor int64
}

func (p *position) Reset() {
	p.mu.Lock()
	p.cursor, p.maxCursor = 0, 0
	p.offset, p.maxOffset = 0, 0
	p.mu.Unlock()
}

func (p *position) AddWindow(i int64) {
	p.mu.Lock()
	p.window += i
	p.mu.Unlock()
}

func (p *position) Window() (c int64) {
	p.mu.RLock()
	c = p.window
	p.mu.RUnlock()
	return
}

func (p *position) MaxOffset() (c int64) {
	p.mu.RLock()
	c = p.maxOffset
	p.mu.RUnlock()
	return
}

func (p *position) MaxCursor() (c int64) {
	p.mu.RLock()
	c = p.maxCursor
	p.mu.RUnlock()
	return
}

func (p *position) Cursor() (c int64) {
	p.mu.RLock()
	c = p.cursor
	p.mu.RUnlock()
	return
}

func (p *position) Offset() (c int64) {
	p.mu.RLock()
	c = p.offset
	p.mu.RUnlock()
	return
}

func (p *position) SetMaxCursor(max int64) {
	p.mu.Lock()
	p.maxCursor = max
	if p.cursor > max {
		p.cursor = max
	}
	p.mu.Unlock()
}

func (p *position) SetMaxOffset(max int64) {
	p.mu.Lock()
	p.maxOffset = max
	if p.offset > max {
		p.offset = max
	}

	p.mu.Unlock()
}

func (p *position) SetCursor(cursor int64) {
	p.mu.Lock()
	p.cursor = cursor
	p.mu.Unlock()
}

func (p *position) SetOffset(offset int64) {
	p.mu.Lock()
	p.offset = offset
	p.mu.Unlock()
}

func (p *position) AddCursor(cursor int64) {
	p.mu.Lock()
	p.addCursor(cursor)
	p.mu.Unlock()
}

func (p *position) AddOffset(offset int64) {
	p.mu.Lock()
	p.addOffset(offset)
	p.mu.Unlock()
}

func (p *position) Add(cursor, offset int64) {
	p.mu.Lock()
	p.addCursor(cursor)
	p.addOffset(offset)
	p.mu.Unlock()
}

func (p *position) addOffset(offset int64) {
	new := p.offset + offset
	if new < 0 {
		p.offset = 0
	} else if new > p.maxOffset {
		p.offset = p.maxOffset
	} else {
		p.offset = new
	}
}

func (p *position) addCursor(cursor int64) {
	new := p.cursor + cursor
	if new < 0 {
		p.cursor = 0
	} else if new > p.maxCursor {
		p.cursor = p.maxCursor
	} else {
		p.cursor = new
	}
}
