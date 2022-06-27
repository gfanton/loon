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
	f.muPosition.Unlock()
}

func (f *FileComponent) Follow(n int) {
	f.muPosition.Lock()
	if !f.lock {
		f.y -= n
	}

	f.muPosition.Unlock()
}

func (f *FileComponent) CursorAdd(y int) {
	f.muPosition.Lock()
	oldy, newy := f.y, f.y+y
	if oldy == 0 && newy > 0 {
		f.lock = true
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

func (f *FileComponent) moveBufferCursor() (cursor int) {
	// if lines := f.buffer.Lines(); lines > 0 {
	if f.y < 0 {
		f.w.Move(f.y)
		f.y = 0
	} else if size := f.w.Size(); f.y > size {
		f.w.Move(f.y - size)
		f.y = size
		// f.buffer.MoveCursor(int64(f.y-height), int64(height))
	}

	return f.y
}

func (f *FileComponent) Redraw(x, y, width, height int) {
	f.muPosition.Lock()
	cursor := f.moveBufferCursor()
	offset := f.updateCursorX(100)

	f.w.DoFork(func(y int, v interface{}) {
		var node *Node
		if v != nil {
			node = v.(*Node)
		}

		if node == nil {
			fillUpLine(f.printer, x, y, width)
		} else {
			node.Line.Print(f.printer, x+1, y, width, int(offset))
		}

		var border rune
		if y == cursor {
			border = '>'
			off := fmt.Sprintf(" -- cursor: %d, yline: %d",
				cursor, y)
			f.printer.Print(width-len(off), y, tcell.StyleDefault, off)
		} else {
			border = ' '
		}
		f.printer.Print(x, y, tcell.StyleDefault, string(border))
	})

	f.muPosition.Unlock()

}

// unc (f *FileComponent) Redraw(x, y, width, height int) {
// 	f.muPosition.Lock()
// 	f.moveBufferCursor()
// 	f.muPosition.Unlock()

// 	input := f.input.Get()
// 	match := f.buffer.FilterLines(input, height)

// 	if f.y > 0 && match.size > f.psize {
// 		f.y += match.size - f.psize
// 	}
// 	f.psize = match.size

// 	if f.y < 0 {
// 		f.y = 0
// 	} else if s := match.size - 1; f.y > s {
// 		f.y = s
// 	}

// 	cursor := f.y + (height - match.size)

// 	offset := f.updateCursorX(match.maxoffset)

// 	// start := match.size
// 	for i := 0; i < height; i++ {
// 		line := match.lines[i]

// 		yline := y + match.size - i - 1
// 		if line == nil {
// 			posy := y + i
// 			fillUpLine(f.printer, x, posy, width)
// 		} else {
// 			line.Print(f.printer, x+1, yline, width, int(offset))
// 		}

// 		posy := y + height - i - 1
// 		var border rune
// 		if i == cursor {
// 			border = '>'
// 			off := fmt.Sprintf(" -- cursor: %d, yline: %d",
// 				cursor, yline)
// 			f.printer.Print(width-len(off), posy, tcell.StyleDefault, off)
// 		} else {
// 			border = ' '
// 		}
// 		f.printer.Print(x, posy, tcell.StyleDefault, string(border))
// 	}

// 	f.muPosition.Unlock()

// start := match.size
// for i, line := range match.lines {
// 	if line == nil {
// 		break
// 	}

// 	yline := start - i
// 	posy := y + yline - 1
// 	line.Print(f.printer, x+1, posy, width, int(offset))
// 	if i == int(cursor) {
// 		f.printer.Print(x, posy, tcell.StyleDefault, ">")

// 		// debug
// 		// off := fmt.Sprintf(" -- cursor: %d, yline: %d, posy: %d",
// 		// 	cursor, yline, posy)
// 		// f.printer.Print(width-len(off), posy, tcell.StyleDefault, off)
// 	}

// }
// 	// debug
// 	// off := fmt.Sprintf(" -- cursor: %d, yline: %d, posy: %d",
// 	// 	cursor, yline, posy)
// 	// f.printer.Print(width-len(off), posy, tcell.StyleDefault, off)
// }
// }
