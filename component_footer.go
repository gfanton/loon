package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type FooterComponent struct {
	s       tcell.Screen
	printer Printer
	buffer  *BufferWindowLine
}

func NewFooterComponent(lcfg *LoonConfig, s tcell.Screen, p Printer, buffer *BufferWindowLine) *FooterComponent {
	return &FooterComponent{
		s:       s,
		printer: p,
		buffer:  buffer,
	}
}

func (i *FooterComponent) Redraw(x, y, width, height int) {
	w, h := i.s.Size()

	lines := i.buffer.Lines()
	line := fmt.Sprintf("height: %d, width: %d, lines: %d", h, w, lines)

	xoffset := i.printer.Print(x, y, tcell.StyleDefault, line)
	fillUpLine(i.printer, xoffset, y, width, tcell.StyleDefault)
}
