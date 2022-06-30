package main

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type FileComponent struct {
	input  *Input
	buffer *Buffer

	w *BufferWindow

	printer Printer

	muPosition  sync.RWMutex
	lock        bool
	x, y, psize int
	maxOffsetX  int
}

func NewFileComponent(print Printer, in *Input, w *BufferWindow) *FileComponent {
	return &FileComponent{
		printer: print,
		input:   in,
		w:       w,
		lock:    true,
	}
}

func (f *FileComponent) ResetPosition() {
	f.muPosition.Lock()
	length, _ := f.w.Size()
	f.w.Reset()
	f.x, f.y = 0, length
	f.muPosition.Unlock()
}

func (f *FileComponent) Follow(n int) {
	f.muPosition.Lock()
	if f.lock {
		f.y += n
	}
	f.muPosition.Unlock()
}

func (f *FileComponent) CursorAdd(y int) {
	f.muPosition.Lock()
	f.y += y
	f.muPosition.Unlock()
}

func (f *FileComponent) OffsetAdd(x int) {
	f.muPosition.Lock()
	f.x += x
	f.muPosition.Unlock()
}

func (f *FileComponent) MaxOffset() (x int) {
	f.muPosition.RLock()
	x = f.maxOffsetX
	f.muPosition.RUnlock()
	return
}

func (f *FileComponent) OffsetSet(x int) {
	f.muPosition.Lock()
	if x > f.maxOffsetX {
		x = f.maxOffsetX
	}
	f.x = x
	f.muPosition.Unlock()
}

func (f *FileComponent) MoveAdd(x, y int) {
	f.muPosition.Lock()
	f.x, f.y = f.x+x, f.y+y
	f.muPosition.Unlock()
}

func (f *FileComponent) updateCursorX(max int) (offset int) {
	if f.x < 0 {
		f.x = 0
	} else if max = max - 1; f.x > max {
		f.x = max
	}

	f.maxOffsetX = max
	offset = f.x
	return
}

func (f *FileComponent) moveBufferCursor(height int) (cursor int) {
	if f.y <= 0 {
		f.w.Move(-f.y)
		f.y = 0
	} else if f.y >= height {
		f.w.Move(height - f.y)
		f.y = height
	}

	switch {
	case !f.lock && f.y == height:
		f.lock = true
	case f.lock && f.y < height:
		f.lock = false
	}

	return f.y
}

func (f *FileComponent) Redraw(x, y, width, height int) {
	f.muPosition.Lock()

	maxsize := height - 1
	winsize, winlen := f.w.Size()
	if winlen != maxsize {
		f.w.Resize(maxsize)
	}

	if winsize < maxsize {
		maxsize = winsize - 1
	}

	cursor := f.moveBufferCursor(maxsize)
	offset := f.updateCursorX(100)

	f.w.DoFork(func(index int, v interface{}) {
		var node *Node
		if v != nil {
			node = v.(*Node)
		}

		indexy := index + y
		icursor := height - indexy
		if node == nil {
			fillUpLine(f.printer, x, indexy, width)
		} else {
			node.Line.Print(f.printer, x+1, indexy, width, int(offset))

		}

		var border rune
		if index == cursor {
			border = '>'
			off := fmt.Sprintf(" -- size: %d, lock: %t, i: %d, cursor: %d, yline: %d", winlen, f.lock, icursor, cursor, index)
			f.printer.Print(width-len(off), indexy, tcell.StyleDefault, off)
		} else {
			border = ' '
		}
		f.printer.Print(x, indexy, tcell.StyleDefault, string(border))
	})

	f.muPosition.Unlock()
}
