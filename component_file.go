package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type FileComponent struct {
	printer Printer

	muWindow sync.Mutex
	window   int64

	input *Input
	pos   *position
	ring  *Ring
}

func NewFileComponent(print Printer, in *Input, pos *position, ring *Ring) *FileComponent {
	return &FileComponent{
		printer: print,
		input:   in,
		ring:    ring,
		pos:     pos,
	}
}

func (f *FileComponent) IncWindow(y int64) {
	f.muWindow.Lock()
	f.window += y
	f.muWindow.Unlock()
}

func (f *FileComponent) Redraw(x, y, width, height int) {
	offset, cursor := f.pos.Offset(), f.pos.Cursor()

	f.muWindow.Lock()
	yoffset := cursor - f.window
	if yoffset < 0 {
		f.window += yoffset
	} else if maxsize := yoffset - int64(height) + 1; maxsize > 0 {
		f.window += maxsize
	}
	window := f.window
	f.muWindow.Unlock()

	input := f.input.Get()
	match := f.ring.FindLine(height, f.window, func(line string) bool {
		if len(input) > 0 {
			return strings.Contains(line, string(input))
		}

		return true
	})

	f.pos.SetMaxOffset(int64(match.maxoffset))

	if match.size < height {
		f.pos.SetMaxCursor(int64(match.size))
	} else {
		f.pos.SetMaxCursor(f.ring.Size())
	}

	start := match.size
	pointer := cursor - window

	for i, line := range match.lines {
		if line == nil {
			break
		}

		nline := y + start - i - 1
		if i == int(pointer) && offset == 0 {
			f.printer.Print(x, nline, tcell.StyleDefault, ">")
			off := fmt.Sprintf(" -- yoffset: %d, wy: %d, pointer: %d", yoffset, f.window, pointer)
			f.printer.Print(line.Len()+1, nline, tcell.StyleDefault, off)
		}

		line.Print(f.printer, x+1, nline, width, int(offset))
	}
}
