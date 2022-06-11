package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type position struct {
	mu sync.RWMutex
	s  tcell.Screen

	// muCursor, muOffset   sync.RWMutex
	cursor, offset       int64
	maxOffset, maxCursor int64
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
	if max < p.cursor {
		p.cursor = max
	}
	p.mu.Unlock()
}

func (p *position) SetMaxOffset(max int64) {
	p.mu.Lock()
	p.maxOffset = max
	if max < p.offset {
		p.offset = max
	}

	p.mu.Unlock()
}

func (p *position) addCursor(cursor int64) {
	if new := p.cursor + cursor; new >= 0 && new < p.maxCursor {
		p.cursor = new
	}
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

func (p *position) addOffset(offset int64) {
	if new := p.offset + offset; new >= 0 && new < p.maxOffset {
		p.offset = new
	}
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
	if cursor != 0 {
		p.addCursor(cursor)
	}
	p.addOffset(offset)
	p.mu.Unlock()
}
