package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type FileComponent struct {
	input   *Input
	buffer  *Buffer
	printer Printer

	muPosition sync.RWMutex
	x, y       int
	maxOffsetX int
}

func NewFileComponent(print Printer, in *Input, ring *Buffer) *FileComponent {
	return &FileComponent{
		printer: print,
		input:   in,
		buffer:  ring,
	}
}

func (f *FileComponent) ResetPosition() {
	f.muPosition.Lock()
	f.x, f.y = 0, 0
	f.buffer.ResetCursor()
	f.muPosition.Unlock()
}

func (f *FileComponent) CursorAdd(y int) {
	f.muPosition.Lock()
	oldy, newy := f.y, f.y+y
	if oldy == 0 && newy > 0 {
		f.buffer.Lock()
	}
	f.y = newy
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

func (f *FileComponent) updateBufferCursor(height int) (cursor int) {
	if lines := f.buffer.Lines(); lines > 0 {
		maxoffset := lines - int64(height)

		if f.y < 0 {
			f.buffer.MoveCursor(maxoffset, int64(f.y))
			f.y = 0
		} else if height = height - 1; f.y > height {
			f.buffer.MoveCursor(maxoffset, int64(f.y-height))
			f.y = height
		}
	}

	return f.y
}

func (f *FileComponent) Redraw(x, y, width, height int) {
	f.muPosition.Lock()

	f.updateBufferCursor(height)

	input := f.input.Get()
	match := f.buffer.FilterLines(input, height)

	var diff int
	if f.y != 0 {
		diff := height - match.size
		if diff > 0 && f.y < diff {
			f.y = diff
		}
	}
	cursor := f.y - diff
	offset := f.updateCursorX(match.maxoffset)

	f.muPosition.Unlock()

	start := match.size
	for i, line := range match.lines {
		if line == nil {
			break
		}

		yline := start - i
		posy := y + yline - 1
		line.Print(f.printer, x+1, posy, width, int(offset))
		if i == int(cursor) {
			f.printer.Print(x, posy, tcell.StyleDefault, ">")

			// debug
			// off := fmt.Sprintf(" -- cursor: %d, yline: %d, posy: %d",
			// 	cursor, yline, posy)
			// f.printer.Print(width-len(off), posy, tcell.StyleDefault, off)
		}

	}

}
