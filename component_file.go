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
	lock                    bool
	cursorX, cursorY, psize int
	maxOffsetX              int
}

func NewFileComponent(print Printer, in *Input, bw *BufferWindow[Line]) *FileComponent {
	return &FileComponent{
		printer: print,
		bw:      bw,
		lock:    true,
	}
}

func (f *FileComponent) ResetPosition() {
	f.muPosition.Lock()
	f.cursorX, f.cursorY = 0, 0
	f.muPosition.Unlock()
}

func (f *FileComponent) Follow(n int) {
	f.muPosition.Lock()
	if f.lock {
		f.cursorY -= n
	}
	f.muPosition.Unlock()
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

func (f *FileComponent) OffsetSet(x int) {
	f.muPosition.Lock()
	if x > f.maxOffsetX {
		x = f.maxOffsetX
	}
	f.cursorX = x
	f.muPosition.Unlock()
}

func (f *FileComponent) MoveAdd(x, y int) {
	f.muPosition.Lock()
	f.cursorX, f.cursorY = f.cursorX+x, f.cursorY+y
	f.muPosition.Unlock()
}

func (f *FileComponent) updateCursorX(max int) (offset int) {
	if f.cursorX < 0 {
		f.cursorX = 0
	} else if max = max - 1; f.cursorX > max {
		f.cursorX = max
	}

	f.maxOffsetX = max
	return f.cursorX
}

func (f *FileComponent) moveBufferCursor() (cursor int) {
	f.bw.Move(-f.cursorY)
	f.cursorY = 0
	// switch {
	// case f.cursorY < 0:
	// 	f.bw.Move(-f.cursorY)
	// 	f.cursorY = 0
	// 	f.bw.Lock(false)
	// case f.cursorY > 0:
	// 	cursor := f.cursorY - max
	// 	f.bw.Move(-cursor)
	// 	f.cursorY = max
	// 	f.bw.Lock(true)
	// default:
	// 	f.bw.Lock(true)

	// }

	return f.cursorY
}

// func (f *FileComponent) updateCursor(width, height int) (offx, offy int) {
// 	offy = f.moveBufferCursor(height)
// 	offx = f.updateCursorX(width)
// 	return
// }

func (f *FileComponent) Redraw(x, y, width, height int) {
	if height == 0 || width == 0 {
		return
	}

	f.muPosition.Lock()

	f.bw.Resize(height)

	// offy := f.moveBufferCursor(height - 1)
	offy := f.moveBufferCursor()

	lines := f.bw.Slice()

	if maxc := len(lines) - 1; offy > maxc {
		offy, f.cursorY = maxc, maxc
	}

	var maxc int
	for _, line := range lines {
		if ll := line.Len(); ll > maxc {
			maxc = ll
		}
	}

	offx := f.updateCursorX(maxc - width)

	var i int
	for ; i < len(lines); i++ {
		indexy := i + y
		// rIndexy := height - indexy
		line := lines[i]
		line.Print(f.printer, x, indexy, width, int(offx))

		// border := ' '
		// if offy == rIndexy {
		// 	border = '>'
		// }
		// f.printer.Print(x, indexy, tcell.StyleDefault, string(border))
	}

	// fillup empty lines
	for ; i < height; i++ {
		indexy := i + y
		f.printer.Print(x, indexy, tcell.StyleDefault, "~")
		fillUpLine(f.printer, x+1, indexy, width)
	}

	f.muPosition.Unlock()
}
