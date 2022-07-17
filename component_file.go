package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type FileComponent struct {
	input *Input
	bw    *BufferWindow[Line]

	printer Printer

	muPosition              sync.RWMutex
	cursorX, cursorY, psize int
	maxOffsetX              int
}

func NewFileComponent(print Printer, in *Input, bw *BufferWindow[Line]) *FileComponent {
	return &FileComponent{
		printer: print,
		bw:      bw,
	}
}

func (f *FileComponent) CursorAdd(y int) {
	f.muPosition.Lock()
	f.cursorY += y
	f.muPosition.Unlock()
}

func (f *FileComponent) OffsetAdd(x int) {
	f.muPosition.Lock()
	f.cursorX += x
	f.muPosition.Unlock()
}

func (f *FileComponent) MaxOffset() (x int) {
	f.muPosition.RLock()
	x = f.maxOffsetX
	f.muPosition.RUnlock()
	return
}

func (f *FileComponent) SetMaxOffset(x int) {
	f.muPosition.Lock()
	f.maxOffsetX = x
	f.muPosition.Unlock()
}

func (f *FileComponent) OffsetSet(x int) {
	f.muPosition.Lock()
	f.cursorX = x
	f.muPosition.Unlock()
}

func (f *FileComponent) MoveAdd(x, y int) {
	f.muPosition.Lock()
	f.cursorX, f.cursorY = f.cursorX+x, f.cursorY+y
	f.muPosition.Unlock()
}

func (f *FileComponent) updateCursorX(max int) (offset int) {
	switch {
	case f.cursorX < 0, max < 0:
		f.cursorX = 0
	case f.cursorX > max:
		f.cursorX = max
	}

	return f.cursorX
}

func (f *FileComponent) moveBufferCursor() (cursor int) {
	f.bw.Move(-f.cursorY)
	f.cursorY = 0
	return f.cursorY
}

func (f *FileComponent) Redraw(x, y, width, height int) {
	if height == 0 || width == 0 {
		return
	}

	f.muPosition.Lock()

	f.bw.Resize(height)

	offy := f.moveBufferCursor()
	offx := f.updateCursorX(f.maxOffsetX - width)

	lines := f.bw.Slice()

	if maxc := len(lines) - 1; offy > maxc {
		offy, f.cursorY = maxc, maxc
	}

	size := f.bw.Do(func(i int, l Line) bool {
		indexy := i + y
		line := lines[i]
		line.Print(f.printer, x, indexy, width, int(offx))
		return true
	})

	// fillup empty lines
	for ; size < height; size++ {
		indexy := size + y
		f.printer.Print(x, indexy, tcell.StyleDefault, "~")
		fillUpLine(f.printer, x+1, indexy, width)
	}

	f.muPosition.Unlock()
}
