package main

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type FileComponent struct {
	s tcell.Screen

	muWindow sync.Mutex
	window   int64

	input *Input
	pos   *position
	ring  *Ring
}

func NewFileComponent(s tcell.Screen, in *Input, p *position, ring *Ring) *FileComponent {
	return &FileComponent{
		s:     s,
		input: in,
		ring:  ring,
		pos:   p,
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
		nline := y + start - i - 1
		if i == int(pointer) && offset == 0 {
			emitStr(f.s, x, nline, tcell.StyleDefault, ">")
			// off := fmt.Sprintf(" -- yoffset: %d, wy: %d, pointer: %d", yoffset, f.window, pointer)
			// emitStr(f.s, len(line)+1, nline, tcell.StyleDefault, off)
		}
		if int(offset) < len(line) {
			emitStr(f.s, x+1, nline, tcell.StyleDefault, line[offset:])
		}
	}
}
