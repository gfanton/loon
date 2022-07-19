package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type sourceFile struct {
	name                 string
	colorDark, colorLigh tcell.Color
	file                 File
}

type FileComponent struct {
	input *Input
	bw    *BufferWindow[Line]

	printer Printer

	muPosition              sync.RWMutex
	cursorX, cursorY, psize int
	maxOffsetX              int

	multisources bool
	sources      map[SourceID]*sourceFile
	sourcesize   int
}

func NewFileComponent(lcfg *LoonConfig, print Printer, sources []File, in *Input, bw *BufferWindowLine) *FileComponent {
	smap := make(map[SourceID]*sourceFile)
	var maxNameSize int
	for _, f := range sources {
		name := filepath.Base(f.Path)
		if len(name) > maxNameSize {
			maxNameSize = len(name)
		}

		sf := &sourceFile{
			name: name,
			file: f,
		}

		if lcfg.BgSourceColor {
			sf.colorLigh = f.ID.Color(0.75)
		}

		if lcfg.FgSourceColor {
			sf.colorDark = f.ID.Color(0)
		}

		smap[f.ID] = sf

	}

	formatmask := fmt.Sprintf("%%-%ds | ", maxNameSize)
	for _, s := range smap {
		s.name = fmt.Sprintf(formatmask, s.name)
	}

	return &FileComponent{
		printer:      print,
		bw:           bw,
		multisources: len(sources) > 1,
		sources:      smap,
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

func (f *FileComponent) printSource(sid SourceID, x, y, offset int) (int, int) {
	s := f.sources[sid]

	if offset <= len(s.name) {
		x = f.printer.Print(x, y, tcell.StyleDefault.Foreground(s.colorDark).Background(s.colorLigh), s.name[offset:])
		offset = 0
	}

	return x, offset - len(s.name)
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

		sx, soffset := x, offx
		if f.multisources {
			sx, soffset = f.printSource(line.Source(), x, indexy, offx)
		}

		line.Print(f.printer, sx, indexy, width, soffset)
		return true
	})

	// fillup empty lines
	for ; size < height; size++ {
		indexy := size + y
		f.printer.Print(x, indexy, tcell.StyleDefault, "~")
		fillUpLine(f.printer, x+1, indexy, width, tcell.StyleDefault)
	}

	f.muPosition.Unlock()
}
